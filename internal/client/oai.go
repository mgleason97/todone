package client

import (
	"context"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

type OpenAIClient struct {
	c     *openai.Client
	model string
}

func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required to start TODOne")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{
		c: &client,
		// Hard-coding model for simplicity though better to take as input to client methods
		model: openai.ChatModelGPT4_1Mini,
	}
}

type GetResponseInput struct {
	SystemPrompt   string                                  `json:"system_prompt"`
	History        []responses.ResponseInputItemUnionParam `json:"history"`
	Tools          []openai.FunctionDefinitionParam        `json:"tools"`
	ResponseFormat responses.ResponseTextConfigParam       `json:"response_format"`
}

type GetResponseOutput struct {
	Answer    string                               `json:"answer"`
	ToolCalls []responses.ResponseFunctionToolCall `json:"tool_calls"`
}

func (o *OpenAIClient) GetResponse(ctx context.Context, req *GetResponseInput) (*GetResponseOutput, error) {
	params := responses.ResponseNewParams{
		Model:        o.model,
		Instructions: param.NewOpt(req.SystemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: req.History,
		},
		Tools: toToolParams(req.Tools),
		Text:  req.ResponseFormat,
	}

	res, err := o.c.Responses.New(ctx, params)
	if err != nil {
		return nil, err
	}

	answer := ""
	var toolCalls []responses.ResponseFunctionToolCall
	for _, out := range res.Output {
		if out.Type == "function_call" {
			fc := out.AsFunctionCall()
			toolCalls = append(toolCalls, fc)
		}
		if out.Type == "message" {
			answer = out.AsMessage().Content[0].Text
		}
	}

	return &GetResponseOutput{
		Answer:    answer,
		ToolCalls: toolCalls,
	}, nil
}

// toToolParams converts functions to tools
func toToolParams(tools []openai.FunctionDefinitionParam) []responses.ToolUnionParam {
	if len(tools) == 0 {
		return nil
	}
	params := make([]responses.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		params = append(params, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
				Strict:      t.Strict,
			},
		})
	}
	return params
}

func UserMessage(msg string) responses.ResponseInputItemUnionParam {
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

func AssistantMessage(msg string) responses.ResponseInputItemUnionParam {
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
