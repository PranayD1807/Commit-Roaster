package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PranayD1807/Commit-Roaster/internal/config"
)

// CmdAI dispatches ai subcommands.
func CmdAI() {
	if len(os.Args) < 3 {
		ShowAIStatus()
		return
	}

	sub := os.Args[2]
	switch sub {
	case "status":
		ShowAIStatus()
	case "enable":
		RunAIEnable()
	case "disable":
		RunAIDisable()
	default:
		fmt.Fprintf(os.Stderr, "  ❌ Unknown AI subcommand: %s\n", sub)
		fmt.Println("     Usage: commit-roaster ai [status|enable|disable]")
		os.Exit(1)
	}
}

func ShowAIStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  %s AI Roaster Status\n", Bold("🤖"))
	if cfg.AIEnabled {
		fmt.Printf("  Status:    %s\n", Green("ENABLED"))
		fmt.Printf("  Provider:  %s\n", config.ProviderDisplayName(cfg.Provider))
		fmt.Printf("  Model:     %s\n", cfg.Model)
		fmt.Printf("  API Key:   %s\n", config.MaskKey(cfg.APIKey))
	} else {
		fmt.Printf("  Status:    %s (defaulting to rule-based roasting)\n", Dim("DISABLED"))
	}
	fmt.Println()
}

func RunAIEnable() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  🤖 Setup AI Roaster Mode")
	fmt.Println()

	// 1. Select provider
	fmt.Println("  Select AI Provider:")
	fmt.Println("    [1] Google Gemini")
	fmt.Println("    [2] Claude")
	fmt.Println("    [3] ChatGPT")
	fmt.Print("  Choose provider (1-3): ")

	provChoice, _ := reader.ReadString('\n')
	provChoice = strings.TrimSpace(provChoice)

	var provider string
	switch provChoice {
	case "1":
		provider = "gemini"
	case "2":
		provider = "claude"
	case "3":
		provider = "chatgpt"
	default:
		fmt.Fprintln(os.Stderr, "  ❌ Invalid choice. Setup aborted.")
		os.Exit(1)
	}

	// 2. Enter API key first so we can use it to fetch models
	fmt.Println()
	fmt.Print("  Enter API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "  ❌ API Key cannot be empty. Setup aborted.")
		os.Exit(1)
	}

	// 3. Select model (dynamic list with local fallback)
	fmt.Println()
	fmt.Printf("  🔄 Fetching live available models for %s...\n", config.ProviderDisplayName(provider))
	models, err := fetchModelsFromAPI(provider, apiKey)
	
	if err != nil {
		fmt.Printf("  ⚠️  Could not fetch live models (%v).\n", err)
		fmt.Println("     Falling back to pre-configured local default models.")
		models = config.ProviderModels(provider)
	}

	fmt.Println()
	fmt.Printf("  Select Model for %s:\n", config.ProviderDisplayName(provider))
	for i, m := range models {
		fmt.Printf("    [%d] %s\n", i+1, m)
	}
	fmt.Printf("    [%d] Other (enter custom model name)\n", len(models)+1)
	fmt.Printf("  Choose model (1-%d): ", len(models)+1)

	modelChoice, _ := reader.ReadString('\n')
	modelChoice = strings.TrimSpace(modelChoice)

	idx := 0
	_, errScan := fmt.Sscan(modelChoice, &idx)
	if errScan != nil || idx < 1 || idx > len(models)+1 {
		fmt.Fprintln(os.Stderr, "  ❌ Invalid choice. Setup aborted.")
		os.Exit(1)
	}

	var model string
	if idx == len(models)+1 {
		fmt.Print("  Enter custom model name: ")
		customModel, _ := reader.ReadString('\n')
		model = strings.TrimSpace(customModel)
		if model == "" {
			fmt.Fprintln(os.Stderr, "  ❌ Model name cannot be empty. Setup aborted.")
			os.Exit(1)
		}
	} else {
		model = models[idx-1]
	}

	// 4. Save Config
	cfg := config.Config{
		AIEnabled: true,
		Provider:  provider,
		Model:     model,
		APIKey:    apiKey,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  ✅ AI roaster enabled with %s (%s)!\n", config.ProviderDisplayName(provider), model)
	fmt.Println()
}

func fetchModelsFromAPI(provider, apiKey string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var url string
	var req *http.Request
	var err error

	switch provider {
	case "gemini":
		url = "https://generativelanguage.googleapis.com/v1beta/models?key=" + apiKey
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	case "claude":
		url = "https://api.anthropic.com/v1/models"
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
		if err == nil {
			req.Header.Set("x-api-key", apiKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		}
	case "chatgpt":
		url = "https://api.openai.com/v1/models"
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var found []string

	switch provider {
	case "gemini":
		var result struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		for _, m := range result.Models {
			name := strings.TrimPrefix(m.Name, "models/")
			// Filter for gemini models that can be used for text generation
			if strings.HasPrefix(name, "gemini-") {
				found = append(found, name)
			}
		}
	case "claude":
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		for _, m := range result.Data {
			if strings.HasPrefix(m.ID, "claude-") {
				found = append(found, m.ID)
			}
		}
	case "chatgpt":
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		for _, m := range result.Data {
			id := m.ID
			if strings.HasPrefix(id, "gpt-") || strings.HasPrefix(id, "o1-") || strings.HasPrefix(id, "o3-") {
				found = append(found, id)
			}
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no matching text generation models found")
	}

	return found, nil
}

func RunAIDisable() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if !cfg.AIEnabled {
		fmt.Println("\n  ℹ️  AI roaster is already disabled.")
		return
	}

	cfg.AIEnabled = false
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n  ✅ AI roaster has been disabled. Reverted to rule-based roasting.")
}
