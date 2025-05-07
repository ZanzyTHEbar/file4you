package genkithandler

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// Define a simple flow that takes a name and returns a greeting.
var GreetingFlow = genkit.DefineFlow("greetingFlow",
	func(ctx context.Context, g *genkit.Genkit, name string) (string, error) {
		model := googlegenai.GoogleAIModel(g, "gemini-1.5-flash") // Or your preferred model

		prompt := "Write a friendly greeting to " + name + "."

		resp, err := genkit.GenerateText(ctx, g, ai.WithModel(model), ai.WithPrompt(prompt))
		if err != nil {
			return "", err
		}

		return resp, nil
	},)