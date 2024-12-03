package main

import "net/url"

var VALID_SCHEMAS = []string{"http", "https"}

// Check if the schema is valid
func isValidSchema(schema string) bool {
	for _, valid := range VALID_SCHEMAS {
		if schema == valid {
			return true
		}
	}
	return false
}

// Check if the host includes a schema
func hasSchema(host string) bool {
	parsed, err := url.Parse(host)
	if err != nil {
		// If it cannot be parsed as a URL, assume there's no schema
		return false
	}
	// If the parsed URL has a schema, it's invalid
	return parsed.Scheme != ""
}
