package airoaster

import (
	"context"
	"fmt"
	"time"
)

const systemPrompt = "You are a witty, snarky, and sarcastic developer. Analyze the following commit message. If it is a good, descriptive, and clean commit message, output exactly 'CLEAN'. Otherwise, generate a short, creative, and biting roast about it (2-3 sentences max). Be funny but direct."

// RoastCommitFn is a variable pointing to the actual RoastCommit function, allowing tests to mock it.
var RoastCommitFn = RoastCommit

// RoastCommit sends a commit message to the chosen AI provider and returns a snarky roast.
func RoastCommit(provider, model, apiKey, commitMsg string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch provider {
	case "gemini":
		return callGemini(ctx, model, apiKey, systemPrompt, commitMsg)
	case "claude":
		return callClaude(ctx, model, apiKey, systemPrompt, commitMsg)
	case "chatgpt":
		return callChatGPT(ctx, model, apiKey, systemPrompt, commitMsg)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}
