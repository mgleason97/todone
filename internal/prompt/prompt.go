package prompt

import "todone/internal/util"

// EnrichPromptWithTask loads a prompt at the given path and appends additional
// information about a task
func EnrichPromptWithTask(path string) string {
	taskDesc := util.MustReadFile("internal/prompt/task_desc.md")
	aggPrompt := util.MustReadFile(path)

	return aggPrompt + "\n" + taskDesc
}
