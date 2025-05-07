package genkithandler

import (
	"context"
	"log"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func InitializeGenkit(ctx context.Context) (*genkit.Genkit, error) {
	// Get the API key from an environment variable.
	// Ensure you have GOOGLE_GENAI_API_KEY or GEMINI_API_KEY set in your environment.
	// For example, in your shell: export GOOGLE_GENAI_API_KEY="YOUR_API_KEY"

	g, err := genkit.Init(ctx,
		genkit.WithPlugins(
			&googlegenai.GoogleAI{}, // Using default constructor, API key will be read from env
		),
	)
	if err != nil {
		return nil, err
	}

	log.Println("Genkit initialized successfully")
	return g, nil
}
