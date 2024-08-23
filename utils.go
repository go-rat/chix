package chix

import (
	"strings"
	"unicode"
)

// quoteEscape escapes the quotes in the string.
func quoteEscape(s string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(s)
}

// parseVendorSpecificContentType check if content type is vendor specific and
// if it is parsable to any known types. If its not vendor specific then returns
// the original content type.
func parseVendorSpecificContentType(cType string) string {
	plusIndex := strings.Index(cType, "+")

	if plusIndex == -1 {
		return cType
	}

	var parsableType string
	if semiColonIndex := strings.Index(cType, ";"); semiColonIndex == -1 {
		parsableType = cType[plusIndex+1:]
	} else if plusIndex < semiColonIndex {
		parsableType = cType[plusIndex+1 : semiColonIndex]
	} else {
		return cType[:semiColonIndex]
	}

	slashIndex := strings.Index(cType, "/")

	if slashIndex == -1 {
		return cType
	}

	return cType[0:slashIndex+1] + parsableType
}

// isASCII checks if the string is ASCII.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
