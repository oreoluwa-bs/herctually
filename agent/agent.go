package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
)

type Agent struct {
	llm            *openai.Client
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

func New(llm *openai.Client, getUserMessage func() (string, bool), tools []ToolDefinition) *Agent {
	return &Agent{
		llm:            llm,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

var systemPrompt = `<role>
   You are 'Herctually,' a sharp, logical, and strategic research assistant, focused on clarity and precise argumentation.
   </role>
   <personality>
   Polite, confident, analytical. You challenge ideas before you support them. You enjoy exposing weak logic. You’re not here to please, you’re here to clarify.
   </personality>
   <rules>
   - Be concise but deep.
   - Separate facts, inferences, and speculation.
   - Always surface what others overlook.
   - Correct errors ruthlessly, including the user’s.
   </rules>
   <goal>
   Deliver insights with precision, not politeness. Your output should feel like a distilled intelligence briefing from someone who actually thinks.
   </goal>`

func (a *Agent) Run(ctx context.Context) error {
	conversation := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	fmt.Println("Chat with Herctually (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}
			userMessage := openai.UserMessage(userInput)
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.Choices[0].Message.ToParam())

		toolResults := []openai.ChatCompletionMessageParamUnion{}

		toolCalls := message.Choices[0].Message.ToolCalls
		for _, call := range toolCalls {
			result := a.executeTool(call.ID, call.Function.Name, []byte(call.Function.Arguments))

			toolResults = append(toolResults, result)
		}

		for _, content := range message.Choices {
			if content.Message.Content != "" {
				fmt.Printf("\u001b[93mHerctually\u001b[0m: %s\n", content.Message.Content)
			}
		}

		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false

		for _, ccmtcu := range toolResults {
			conversation = append(conversation, ccmtcu)
		}
	}
	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	opentools := []openai.ChatCompletionToolUnionParam{}

	for _, tool := range a.tools {
		opentools = append(opentools, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  tool.InputSchema,
				},
			},
		})
	}

	message, err := a.llm.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:               openai.ChatModelGPT4o,
		MaxCompletionTokens: openai.Int(1024),
		Messages:            conversation,
		Tools:               opentools,
	})
	return message, err
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) openai.ChatCompletionMessageParamUnion {
	var toolDef ToolDefinition
	var found bool

	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}

	if !found {
		return openai.ToolMessage("tool not found", id)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return openai.ToolMessage(err.Error(), id)
	}

	return openai.ToolMessage(response, id)
}
