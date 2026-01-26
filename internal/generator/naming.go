package generator

import (
	"strings"
	"unicode"
)

// ToPascalCase converts a string to PascalCase.
// Examples: "user_name" -> "UserName", "userName" -> "UserName"
func ToPascalCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToCamelCase converts a string to camelCase.
// Examples: "user_name" -> "userName", "UserName" -> "userName"
func ToCamelCase(s string) string {
	if s == "" {
		return s
	}

	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}

	// Handle all-caps or leading acronyms
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ToSnakeCase converts a string to snake_case.
// Examples: "UserName" -> "user_name", "userName" -> "user_name"
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToEnumConstName creates an enum constant name.
// Examples: Role + ADMIN -> RoleAdmin, Status + IN_PROGRESS -> StatusInProgress
func ToEnumConstName(enumName, valueName string) string {
	// Convert value name: SOME_VALUE -> SomeValue
	valuePascal := ToPascalCase(strings.ToLower(valueName))
	return enumName + valuePascal
}
