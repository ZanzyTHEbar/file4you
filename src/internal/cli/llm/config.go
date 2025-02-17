package llm

import "github.com/spf13/cobra"

// LLMConfig holds configuration for the LLM agent
type LLMConfig struct {
    Model          string
    Temperature    float64
    MaxTokens      int
    PromptTemplate string
    BatchSize      int
}

// AddLLMFlags adds LLM-specific flags to a command
func AddLLMFlags(cmd *cobra.Command, config *LLMConfig) {
    cmd.Flags().StringVar(&config.Model, "model", "gpt-3.5-turbo", "LLM model to use")
    cmd.Flags().Float64Var(&config.Temperature, "temperature", 0.7, "Model temperature (0-1)")
    cmd.Flags().IntVar(&config.MaxTokens, "max-tokens", 2048, "Maximum tokens per request")
    cmd.Flags().StringVar(&config.PromptTemplate, "prompt-template", "", "Custom prompt template file")
    cmd.Flags().IntVar(&config.BatchSize, "batch-size", 10, "Number of files to process in parallel")
}