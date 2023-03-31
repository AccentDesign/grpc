package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

// Validator struct
type Validator struct{}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{}
}

// IsEmpty checks if the given string is empty
func (v *Validator) IsEmpty(s string) bool {
	return len(s) == 0
}

// Matches checks if the given string matches the given regex pattern
func (v *Validator) Matches(s string, rx *regexp.Regexp) bool {
	return rx.MatchString(s)
}

// IsStringLength checks if the given string's length is between min and max (inclusive)
func (v *Validator) IsStringLength(s string, min, max int) bool {
	length := len(s)
	return length >= min && length <= max
}
