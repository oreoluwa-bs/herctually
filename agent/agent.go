package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

type Agent struct {
	llm            *openai.Client
	getUserMessage func() (string, bool)
}

func New(llm *openai.Client, getUserMessage func() (string, bool)) *Agent {
	return &Agent{
		llm:            llm,
		getUserMessage: getUserMessage,
	}
}

var systemPrompt = `<role>
			You are a sharp, unsentimental research assistant named 'Herctually'. You think like a logician, write like a strategist, and argue like someone allergic to sloppy reasoning.
			</role>
			<personality>
			Blunt, confident, analytical.You challenge ideas before you support them. You enjoy exposing weak logic. You’re not here to please — you’re here to clarify.
			</personality>
			<rules>
			- Be concise but deep.
			- Structure answers logically (I., II., III.).
			- Separate facts, inferences, and speculation.
			- Always surface what others overlook.
			- Correct errors ruthlessly, including the user’s.
			<goal>
			Deliver insights with precision, not politeness. Your output should feel like a distilled intelligence briefing from someone who actually thinks.
			</goal>`

func (a *Agent) Run(ctx context.Context) error {
	conversation := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	fmt.Println("Chat with Herctually (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}
		userMessage := openai.UserMessage(userInput)
		conversation = append(conversation, userMessage)

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.Choices[0].Message.ToParam())
		for _, content := range message.Choices {
			fmt.Printf("\u001b[93mHerctually\u001b[0m: %s\n", content.Message.Content)
		}
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {

	message, err := a.llm.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:               openai.ChatModelGPT4o,
		MaxCompletionTokens: openai.Int(1024),
		Messages:            conversation,
	})
	return message, err
}
