package guardrails

type ValidationResult struct {
	IsValid  bool   // true = allowed ; false = blocked
	Reason   string // Why the query was blocked
	Category string // "toxic", "off_topic", "pii", "prompt_injection"
	Method   string // "static" or claude
}
