package llm

import "fmt"

// DecisionValidator validates DestinationDecision
type DecisionValidator struct {
	MaxPathLength   int
	AllowedFolders  []string
	DisallowedChars []rune
}

func (v *DecisionValidator) Validate(decision DestinationDecision) error {
	if len(decision.DestinationFolder) > v.MaxPathLength {
		return fmt.Errorf("destination folder path too long")
	}

	// Add more validation logic
	return nil
}
