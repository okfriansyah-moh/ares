package arslib

// Manifest is the parsed .ai/manifest.yaml.
type Manifest struct {
	Version  string   `yaml:"version"`
	Project  Project  `yaml:"project"`
	Defaults Defaults `yaml:"defaults,omitempty"`
}

// Project holds repository metadata from manifest.yaml.
type Project struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Repository  string `yaml:"repository,omitempty"`
}

// Defaults holds optional manifest defaults.
type Defaults struct {
	Agent string `yaml:"agent,omitempty"`
}

// Repository is the in-memory representation of a .ai/ directory.
type Repository struct {
	Manifest     Manifest
	Instructions []Instruction
	Agents       []Agent
	Skills       []Skill
	Prompts      []Prompt
}

// Instruction is a repository-wide rule file under instructions/.
type Instruction struct {
	ID      string // filename stem
	Path    string
	Content string
}

// Agent is an agent definition under agents/<id>/AGENT.md.
type Agent struct {
	ID        string // directory name
	Path      string
	Content   string
	SkillRefs []string
}

// ExtraFile is a supporting file within a skill directory (anything other than SKILL.md).
// Rel is the path relative to the skill directory root, using forward slashes.
type ExtraFile struct {
	Rel     string
	Content []byte
}

// Skill is a skill definition under skills/<id>/SKILL.md.
type Skill struct {
	ID         string
	Path       string
	Content    string
	References []string
	ExtraFiles []ExtraFile
}

// Prompt is a prompt template under prompts/.
type Prompt struct {
	ID      string
	Path    string
	Content string
}

// FindingLevel is the severity of a validation finding.
type FindingLevel int

const (
	OK FindingLevel = iota
	Warning
	Error
)

// String returns a human-readable label for the finding level.
func (l FindingLevel) String() string {
	switch l {
	case OK:
		return "OK"
	case Warning:
		return "Warning"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

// Finding is a single validation result.
type Finding struct {
	Level   FindingLevel `json:"level"`
	Path    string       `json:"path"`
	Message string       `json:"message"`
}
