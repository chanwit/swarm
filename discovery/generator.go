package discovery

import (
	"fmt"
	"regexp"
)

//
// This function increases value in the form of String.
// It checks the last character of string 's'
// then returns the value next to s.
//
// For example,
//
//  - "0" will become "1"
//  - "01" will become "02", the leading zero will be preserved
//  - "09" will become "10", the leading zero will be increased to "1"
//  - "9" will become "10", an extra string "1" will be prepended, if overflow
//  - "a0" will become "a1"
//  - "a9" will become "b0", as the last digit will overflow and change "a" to "b"
//  - "az" will become "ba"
//  - "AZ" will become "BA"
//
func incString(s string) string {
	if s == "" {
		return "1"
	}
	lastIndex := len(s) - 1
	lastChar := s[lastIndex]
	switch lastChar {
	case '9':
		return incString(s[:lastIndex]) + "0"
	case 'Z':
		return incString(s[:lastIndex]) + "A"
	case 'z':
		return incString(s[:lastIndex]) + "a"
	default:
		return s[:lastIndex] + string(lastChar+1)
	}
}

//
// IP and Hostname generator
//
func Generate(pattern string) []string {
	re, _ := regexp.Compile(`\[(.+):(.+)\]`)
	submatch := re.FindStringSubmatch(pattern)
	if submatch == nil {
		return []string{pattern}
	}

	from := submatch[1]
	to := submatch[2]

	template := re.ReplaceAllString(pattern, "%s")

	result := make([]string, 0)
	for val := from; ; val = incString(val) {
		entry := fmt.Sprintf(template, val)
		result = append(result, entry)
		if val == to {
			break
		}
	}

	return result
}
