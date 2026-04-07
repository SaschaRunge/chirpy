package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			input:    "I really need a kerfuffle to go to bed sooner, Fornax !",
			expected: "I really need a **** to go to bed sooner, **** !",
		},
	}

	for _, c := range cases {
		filteredText := filterText(c.input)
		if c.expected != filteredText {
			t.Errorf("filter mismatch: \nactual: %s\nexpected: %s", filteredText, c.expected)
		}
	}
}
