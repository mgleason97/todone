package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"todone/internal"
	"todone/internal/aggregate"
	"todone/internal/client"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

type Agent struct {
	client       *client.OpenAIClient
	userPipe     <-chan string
	history      []responses.ResponseInputItemUnionParam
	systemPrompt string
	todos        []internal.TODO
	cfg          internal.Config
}

func New(client *client.OpenAIClient, pipe <-chan string, cfg internal.Config) (Agent, error) {
	prompt, err := loadSystemPrompt()
	if err != nil {
		return Agent{}, err
	}
	return Agent{
		client:       client,
		userPipe:     pipe,
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
		case msg, ok := <-a.userPipe:
			if !ok {
				return nil
			}
			if strings.TrimSpace(msg) == "" {
				continue
			}
			a.history = append(a.history, userMessage(msg))
			if err := a.turn(ctx); err != nil {
				return err
			}
		}
	}
}

func (a *Agent) turn(ctx context.Context) error {
	for {
		req := client.GetResponseInput{
			SystemPrompt: a.systemPrompt,
			History:      a.history,
			Tools:        a.tools(),
		}
		res, err := a.client.GetResponse(ctx, &req)
		if err != nil {
			return err
		}

		// No tool calls requested means agent is done with its turn
		if len(res.ToolCalls) == 0 {
			a.history = append(a.history, assistantMessage(res.Answer))
			fmt.Printf("Agent: %s\n", res.Answer)
			return nil
		}

		if err := a.handleToolCalls(res.ToolCalls); err != nil {
			return err
		}
	}
}

func userMessage(msg string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleUser,
			Type: responses.EasyInputMessageTypeMessage,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt(msg),
			},
		},
	}
}

func assistantMessage(msg string) responses.ResponseInputItemUnionParam {
	return responses.ResponseInputItemUnionParam{
		OfMessage: &responses.EasyInputMessageParam{
			Role: responses.EasyInputMessageRoleAssistant,
			Type: responses.EasyInputMessageTypeMessage,
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt(msg),
			},
		},
	}
}

func loadSystemPrompt() (string, error) {
	data, err := os.ReadFile("internal/agent/prompt.md")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
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
			todos, err := aggregate.Aggregate(a.cfg)
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
