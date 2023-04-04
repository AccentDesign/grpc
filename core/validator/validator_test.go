package validator_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/accentdesign/grpc/core/validator"
)

func TestIsEmpty(t *testing.T) {
	v := validator.New()

	testCases := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"not empty", false},
		{"   ", false},
		{"\t\n", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, v.IsEmpty(tc.input))
	}
}

func TestMatches(t *testing.T) {
	v := validator.New()
	pattern := `\d+$`
	rx := regexp.MustCompile(pattern)

	testCases := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"abc", false},
		{"12a", false},
		{"", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, v.Matches(tc.input, rx))
	}
}

func TestIsStringLength(t *testing.T) {
	v := validator.New()

	testCases := []struct {
		input          string
		min, max       int
		expectedResult bool
	}{
		{"hello", 1, 10, true},
		{"world", 5, 5, true},
		{"too long", 1, 5, false},
		{"too short", 10, 20, false},
		{"", 0, 0, true},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, v.IsStringLength(tc.input, tc.min, tc.max))
	}
}

func TestEmailValidation(t *testing.T) {
	v := validator.New()

	testCases := []struct {
		input    string
		expected bool
	}{
		{"john.doe@example.com", true},
		{"john.doe@subdomain.example.com", true},
		{"john.doe@example.co.uk", true},
		{"john.doe@sub_domain.example.com", false},
		{"john.doe@example", false},
		{"john.doe@example.", false},
		{"john.doe@", false},
		{"john.doe.example", false},
		{"", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, v.Matches(tc.input, validator.EmailRX))
	}
}
