package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

	// 2. Select model
	models := config.ProviderModels(provider)
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
	_, err := fmt.Sscan(modelChoice, &idx)
	if err != nil || idx < 1 || idx > len(models)+1 {
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

	// 3. Enter API key
	fmt.Println()
	fmt.Print("  Enter API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "  ❌ API Key cannot be empty. Setup aborted.")
		os.Exit(1)
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
