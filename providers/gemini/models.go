package gemini

// Models returns the list of known Gemini model identifiers.
func Models() []string {
	return []string{
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
		"gemini-1.5-flash",
		"gemini-1.5-pro",
		"gemini-2.5-pro-preview-05-06",
		"gemini-2.5-flash-preview-04-17",
	}
}
