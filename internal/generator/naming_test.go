package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"user", "User"},
		{"user_name", "UserName"},
		{"userName", "UserName"},
		{"user-name", "UserName"},
		{"user name", "UserName"},
		{"USER_NAME", "USERNAME"},
		{"id", "Id"},
		{"ID", "ID"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"User", "user"},
		{"UserName", "userName"},
		{"user_name", "userName"},
		{"ID", "iD"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"User", "user"},
		{"UserName", "user_name"},
		{"userName", "user_name"},
		{"ID", "i_d"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToEnumConstName(t *testing.T) {
	tests := []struct {
		enumName  string
		valueName string
		expected  string
	}{
		{"Role", "ADMIN", "RoleAdmin"},
		{"Role", "USER", "RoleUser"},
		{"Status", "IN_PROGRESS", "StatusInProgress"},
		{"Priority", "HIGH", "PriorityHigh"},
	}

	for _, tt := range tests {
		t.Run(tt.enumName+"_"+tt.valueName, func(t *testing.T) {
			result := ToEnumConstName(tt.enumName, tt.valueName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
