package yelp

import (
	"testing"
)

func TestPercentEncode(t *testing.T) {
	toAttempt := []string{"",
		"abc",
		"abc abc",
		"a",
		"abc!abc@abc",
		"abc#abc$abc",
		"abc%abc^abc",
		"abc&abc*abc",
		"abc(abc)abc",
		"abc-abc_abc",
		"abc=abc+abc"}
	expected := []string{"",
		"abc",
		"abc%20abc",
		"a",
		"abc%21abc%40abc",
		"abc%23abc%24abc",
		"abc%25abc%5Eabc",
		"abc%26abc%2Aabc",
		"abc%28abc%29abc",
		"abc-abc_abc",
		"abc%3Dabc%2Babc"}
		
	if len(toAttempt) != len(expected) {
		t.Errorf("Invalid test data supplied: %d attempts unequal to %d expected results", len(toAttempt), len(expected))
	}

	for i := range toAttempt {
		if percentEncode(toAttempt[i]) != expected[i] {
			t.Errorf("percentEncode('%s') was not equal to '%s'", toAttempt[i], expected[i])
		}
	}
}