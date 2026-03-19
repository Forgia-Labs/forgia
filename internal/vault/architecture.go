package vault

// Architecture represents the C4 Level 1-2 system design.
type Architecture struct {
	SystemContext       *SystemContext       `yaml:"system_context"`
	Containers          *Containers          `yaml:"containers"`
	TechnologyDecisions *TechnologyDecisions `yaml:"technology_decisions"`
	QualityAttributes   *QualityAttributes   `yaml:"quality_attributes"`
	Constraints         *ArchConstraints     `yaml:"constraints"`
	Glossary            []GlossaryTerm       `yaml:"glossary,omitempty"`
}

// SystemContext is C4 Level 1 — the system in its environment.
type SystemContext struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Actors      []Actor  `yaml:"actors,omitempty"`
	External    []string `yaml:"external_systems,omitempty"`
	Boundaries  struct {
		Inside  []string `yaml:"inside"`
		Outside []string `yaml:"outside"`
	} `yaml:"boundaries"`
}

// Actor is a user or external system that interacts with the platform.
type Actor struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // human, system, service
	Role string `yaml:"role"`
}

// Containers is C4 Level 2 — services, apps, databases.
type Containers struct {
	Items []Container `yaml:"containers"`
}

// Container is a deployable unit (service, app, database).
type Container struct {
	Name       string `yaml:"name"`
	Technology string `yaml:"technology"`
	Purpose    string `yaml:"purpose"`
	Context    string `yaml:"context,omitempty"` // maps to bounded context
}

// TechnologyDecisions is the ADR master — choices that apply everywhere.
type TechnologyDecisions struct {
	Decisions []TechDecision `yaml:"decisions"`
}

// TechDecision records a technology choice with rationale.
type TechDecision struct {
	Area      string `yaml:"area"`
	Choice    string `yaml:"choice"`
	Rationale string `yaml:"rationale"`
	ADR       string `yaml:"adr,omitempty"` // link to ADR file
}

// QualityAttributes defines non-functional requirements.
type QualityAttributes struct {
	Latency      string `yaml:"latency,omitempty"`
	Availability string `yaml:"availability,omitempty"`
	Security     string `yaml:"security,omitempty"`
	Scalability  string `yaml:"scalability,omitempty"`
}

// ArchConstraints captures budget, team, timeline.
type ArchConstraints struct {
	Budget   string   `yaml:"budget,omitempty"`
	Team     string   `yaml:"team,omitempty"`
	Timeline string   `yaml:"timeline,omitempty"`
	Other    []string `yaml:"other,omitempty"`
}

// GlossaryTerm defines a ubiquitous language term (DDD).
type GlossaryTerm struct {
	Term    string `yaml:"term"`
	Meaning string `yaml:"meaning"`
}

// BoundedContext represents a DDD bounded context.
type BoundedContext struct {
	Name             string              `yaml:"name"`
	Description      string              `yaml:"description"`
	Responsibilities []string            `yaml:"responsibilities,omitempty"`
	Language         []GlossaryTerm      `yaml:"ubiquitous_language,omitempty"`
	Interfaces       []ContextInterface  `yaml:"interfaces,omitempty"`
	Dependencies     ContextDependencies `yaml:"dependencies"`
	Status           ContextStatus       `yaml:"status"`

	// Internal.
	FilePath string `yaml:"-"`
}

// ContextInterface defines a contract between bounded contexts.
type ContextInterface struct {
	With     string `yaml:"with"`
	Protocol string `yaml:"protocol"`
	Contract string `yaml:"contract"`
}

// ContextDependencies lists what depends on what.
type ContextDependencies struct {
	DependsOn   []string `yaml:"depends_on,omitempty"`
	DependedBy  []string `yaml:"depended_by,omitempty"`
}

// ContextStatus tracks implementation progress.
type ContextStatus struct {
	FDCreated   bool `yaml:"fd_created"`
	FDCompleted bool `yaml:"fd_completed"`
}
