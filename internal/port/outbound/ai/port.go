// Package ai defines outbound port interfaces for AI / NLP providers.
package ai

import "context"

// IntentResult holds the structured output from an intent extraction call.
type IntentResult struct {
	Intent     string
	Confidence float64
	Payload    map[string]interface{}
	RawText    string
}

// IntentExtractor extracts structured intent from free-form natural-language text.
type IntentExtractor interface {
	// Extract parses text in the given language (e.g. "id" for Indonesian)
	// and returns the detected intent.
	Extract(ctx context.Context, text, lang string) (*IntentResult, error)
}
