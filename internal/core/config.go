package core

// SectionConfig defines the mapping of task categories to section names
type SectionConfig struct {
	Overdue         string   `json:"overdue"`
	DueToday        string   `json:"due_today"`
	DueThisWeek     string   `json:"due_this_week"`
	DueLater        string   `json:"due_later"`
	NoDate          string   `json:"no_date"`
	IgnoredSections []string `json:"ignored_sections,omitempty"`
}

// DefaultSectionConfig returns the default section configuration
func DefaultSectionConfig() SectionConfig {
	return SectionConfig{
		Overdue:         "Overdue",
		DueToday:        "Due today",
		DueThisWeek:     "Due within the next 7 days",
		DueLater:        "Due later",
		NoDate:          "Recently assigned",
		IgnoredSections: []string{},
	}
}