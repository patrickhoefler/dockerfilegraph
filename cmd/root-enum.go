package cmd

import (
	"errors"
	"sort"
)

// enum is a Cobra-compatible wrapper for defining a string flag
// that can be one of a specified set of values.
type enum struct {
	allowedValues []string
	value         string
}

// newEnum returns a new enum flag.
func newEnum(defaultValue string, allowedValues ...string) *enum {
	allowedValues = append(allowedValues, defaultValue)
	sort.Strings(allowedValues)

	return &enum{
		allowedValues: allowedValues,
		value:         defaultValue,
	}
}

// String returns a string representation of the enum flag.
func (e *enum) String() string { return e.value }

// Set assigns the provided string to the enum receiver.
// It returns an error if the string is not an allowed value.
func (e *enum) Set(s string) error {
	for _, val := range e.allowedValues {
		if val == s {
			e.value = s
			return nil
		}
	}

	return errors.New("invalid value: " + s)
}

// Type returns a string representation of the enum type.
func (e *enum) Type() string { return "" }

// AllowedValues returns a slice of the flag's valid values.
func (e *enum) AllowedValues() []string {
	return e.allowedValues
}
