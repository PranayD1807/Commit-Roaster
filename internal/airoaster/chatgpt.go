package airoaster

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	openaiOption "github.com/openai/openai-go/option"
)

func callChatGPT(ctx context.Context, model, apiKey, systemPrompt, commitMsg string) (string, error) {
	client := openai.NewClient(
		openaiOption.WithAPIKey(apiKey),
	)

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:               openai.ChatModel(model),
		MaxCompletionTokens: openai.Int(200),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(commitMsg),
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response choices returned from ChatGPT")
}
