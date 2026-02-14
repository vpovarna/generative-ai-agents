package rewrite

import (
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
)

type Rewriter struct {
	claudeClient bedrock.Client
}
