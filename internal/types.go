package internal

type Config struct {
	Repos     []Repo          `toml:"repos"`
	Messaging MessagingConfig `toml:"messaging"`
}

type Repo struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

type MessagingConfig struct {
	App      string    `toml:"app"`
	Contacts []Contact `toml:"contacts"`
}

type Contact struct {
	Name string `toml:"name"`
}

// TODO represents a single actionable item aggregated from any source.
// Priority: 0 = highest, 2 = lowest. EffortMinutes is a best-effort estimate.
type TODO struct {
	Title         string
	Description   string
	EffortMinutes int
	Priority      int
}

type SourceType string

const (
	SourceTypeCode    SourceType = "code"
	SourceTypeMessage SourceType = "message"
)

type RawTask struct {
	Source       SourceType `json:"source"`
	TaskMetadata any        `json:"metadata"`
}

type Task struct {
	Source        SourceType
	Title         string
	Description   string
	EffortMinutes int
	Priority      int
}
