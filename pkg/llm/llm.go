package llm

// LLM is the minimal interface any model client must implement to be used by the bot
type LLM interface {
	Ask(prompt string) (string, error)
}
