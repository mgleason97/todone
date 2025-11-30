package prompt

import _ "embed"

var (
	//go:embed agent.md
	agentPrompt string
	//go:embed task_desc.md
	taskDesc string
	//go:embed code_enrichment.md
	codeEnrichment string
)

var AgentPrompt = agentPrompt + "\n" + taskDesc
var CodeEnrichmentPrompt = codeEnrichment + "\n" + taskDesc
