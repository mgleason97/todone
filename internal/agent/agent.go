package agent

import (
	"context"
	"encoding/json"
	"strings"
	"todone/internal"
	"todone/internal/aggregate"
	"todone/internal/client"
	"todone/internal/prompt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

type Agent struct {
	client       *client.OpenAIClient
	userIn       <-chan string
	agentOut     chan<- string
	history      []responses.ResponseInputItemUnionParam
	systemPrompt string
	todos        []internal.TODO
	cfg          internal.Config
}

func New(client *client.OpenAIClient, userIn <-chan string, agentOut chan<- string, cfg internal.Config) (Agent, error) {
	prompt := prompt.EnrichPromptWithTask("internal/prompt/agent.md")
	return Agent{
		client:       client,
		userIn:       userIn,
		agentOut:     agentOut,
		history:      []responses.ResponseInputItemUnionParam{},
		systemPrompt: prompt,
		cfg:          cfg,
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-a.userIn:
			if !ok {
				return nil
			}
			msg = strings.TrimSpace(msg)
			if msg == "" {
				continue
			}
			a.history = append(a.history, client.UserMessage(msg))

			ans, err := a.turn(ctx)
			if err != nil {
				return err
			}
			a.agentOut <- ans
		}
	}
}

func (a *Agent) turn(ctx context.Context) (string, error) {
	for {
		req := client.GetResponseInput{
			SystemPrompt: a.systemPrompt,
			History:      a.history,
			Tools:        a.tools(),
		}
		res, err := a.client.GetResponse(ctx, &req)
		if err != nil {
			return "", err
		}

		// No tool calls requested means agent is done with its turn
		if len(res.ToolCalls) == 0 {
			a.history = append(a.history, client.AssistantMessage(res.Answer))
			return res.Answer, nil
		}

		if err := a.handleToolCalls(res.ToolCalls); err != nil {
			return "", err
		}
	}
}

func (a *Agent) tools() []openai.FunctionDefinitionParam {
	return []openai.FunctionDefinitionParam{
		{
			Name:        "aggregate_todos",
			Description: param.NewOpt("Aggregate TODOs from configured repositories and messaging sources."),
		},
	}
}

func (a *Agent) handleToolCalls(calls []responses.ResponseFunctionToolCall) error {
	for _, call := range calls {
		a.history = append(a.history, responses.ResponseInputItemParamOfFunctionCall(call.Arguments, call.CallID, call.Name))

		switch call.Name {
		case "aggregate_todos":
			todos, err := aggregate.Aggregate(a.client, a.cfg)
			if err != nil {
				return err
			}
			a.todos = todos
			payload, err := json.Marshal(todos)
			if err != nil {
				return err
			}
			a.history = append(a.history, responses.ResponseInputItemParamOfFunctionCallOutput(call.CallID, string(payload)))
		default:
			a.history = append(a.history, responses.ResponseInputItemParamOfFunctionCallOutput(call.CallID, "unknown tool"))
		}
	}
	return nil
}
