package ginkgo_test

import (
	"testing"

	"github.com/itsLeonB/ginkgo"
	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	// Test that all error constants have the expected values
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "ErrMissingToken",
			constant: ginkgo.ErrMissingToken,
			expected: "missing token",
		},
		{
			name:     "ErrInvalidToken",
			constant: ginkgo.ErrInvalidToken,
			expected: "invalid token",
		},
		{
			name:     "ErrUserNotFound",
			constant: ginkgo.ErrUserNotFound,
			expected: "user not found",
		},
		{
			name:     "ErrNoPermission",
			constant: ginkgo.ErrNoPermission,
			expected: "user does not have the required permission",
		},
		{
			name:     "ErrTypeUnauthorized",
			constant: ginkgo.ErrTypeUnauthorized,
			expected: "Unauthorized",
		},
		{
			name:     "ErrTypeForbidden",
			constant: ginkgo.ErrTypeForbidden,
			expected: "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestErrorConstantsNotEmpty(t *testing.T) {
	// Test that no error constants are empty strings
	constants := []struct {
		name  string
		value string
	}{
		{"ErrMissingToken", ginkgo.ErrMissingToken},
		{"ErrInvalidToken", ginkgo.ErrInvalidToken},
		{"ErrUserNotFound", ginkgo.ErrUserNotFound},
		{"ErrNoPermission", ginkgo.ErrNoPermission},
		{"ErrTypeUnauthorized", ginkgo.ErrTypeUnauthorized},
		{"ErrTypeForbidden", ginkgo.ErrTypeForbidden},
	}

	for _, constant := range constants {
		t.Run(constant.name, func(t *testing.T) {
			assert.NotEmpty(t, constant.value, "%s should not be empty", constant.name)
		})
	}
}

func TestErrorConstantsUniqueness(t *testing.T) {
	// Test that all error constants have unique values
	constants := map[string]string{
		"ErrMissingToken":     ginkgo.ErrMissingToken,
		"ErrInvalidToken":     ginkgo.ErrInvalidToken,
		"ErrUserNotFound":     ginkgo.ErrUserNotFound,
		"ErrNoPermission":     ginkgo.ErrNoPermission,
		"ErrTypeUnauthorized": ginkgo.ErrTypeUnauthorized,
		"ErrTypeForbidden":    ginkgo.ErrTypeForbidden,
	}

	// Create a reverse map to check for duplicates
	valueToName := make(map[string][]string)
	for name, value := range constants {
		valueToName[value] = append(valueToName[value], name)
	}

	// Check that no value appears more than once
	for value, names := range valueToName {
		if len(names) > 1 {
			t.Errorf("Duplicate error constant value '%s' found in: %v", value, names)
		}
	}
}
