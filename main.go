package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/oreoluwa-bs/the-gains-father/agent"
)

func main() {

	apiKey := os.Getenv("OPENROUTER_APIKEY")
	if apiKey == "" {
		log.Panic("OPENROUTER_APIKEY environment variable is required")
		os.Exit(1)
	}

	baseURL := os.Getenv("OPENROUTER_BASEURL")
	if baseURL == "" {
		log.Panic("OPENROUTER_BASEURL environment variable is required")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}

		return scanner.Text(), true
	}

	llm := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	ag := agent.New(&llm, getUserMessage)

	if err := ag.Run(context.Background()); err != nil {
		fmt.Printf("Error: %s\n", err)
	}

}
