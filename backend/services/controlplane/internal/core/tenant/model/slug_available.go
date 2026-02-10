package tenantmodel

type SlugAvailable struct {
	Available   bool
	Suggestions []string
}

func NewSlugAvailable(available bool, suggestions []string) *SlugAvailable {
	return &SlugAvailable{
		Available:   available,
		Suggestions: suggestions,
	}
}

func (s *SlugAvailable) GetIsAvailable() bool {
	return s.Available
}

func (s *SlugAvailable) SetIsAvailable(available bool) {
	if s == nil {
		return
	}
	s.Available = available
}

func (s *SlugAvailable) GetSuggestions() []string {
	return s.Suggestions
}

func (s *SlugAvailable) AddSuggestion(suggestion string) {
	s.Suggestions = append(s.Suggestions, suggestion)
}
