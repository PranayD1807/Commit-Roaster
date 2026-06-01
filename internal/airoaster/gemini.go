package airoaster

import (
	"context"

	"google.golang.org/genai"
)

func callGemini(ctx context.Context, model, apiKey, systemPrompt, commitMsg string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return "", err
	}

	resp, err := client.Models.GenerateContent(ctx, model, genai.Text(commitMsg), &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemPrompt},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}
