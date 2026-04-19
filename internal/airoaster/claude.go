package airoaster

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func callClaude(ctx context.Context, model, apiKey, systemPrompt, commitMsg string) (string, error) {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(200),
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(commitMsg)),
		},
	})
	if err != nil {
		return "", err
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text block returned from Claude")
}
