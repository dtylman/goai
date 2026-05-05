package deepseek

import "github.com/cohesion-org/deepseek-go"

// Models returns the list of known DeepSeek model identifiers.
func Models() []string {
	return []string{
		deepseek.DeepSeekChat,
		deepseek.DeepSeekReasoner,
	}
}
