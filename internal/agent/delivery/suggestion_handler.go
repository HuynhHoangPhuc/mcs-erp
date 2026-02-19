package delivery

import (
	"net/http"
)

// Suggestion represents a contextual action button shown on an entity page.
type Suggestion struct {
	Label   string `json:"label"`
	Message string `json:"message"` // pre-filled chat message when clicked
	Icon    string `json:"icon,omitempty"`
}

// SuggestionHandler handles GET /api/v1/agent/suggestions
type SuggestionHandler struct{}

// NewSuggestionHandler creates a SuggestionHandler.
func NewSuggestionHandler() *SuggestionHandler {
	return &SuggestionHandler{}
}

// suggestionRules maps entity_type to contextual suggestions.
// Rule-based: no LLM call needed — deterministic based on entity type.
var suggestionRules = map[string][]Suggestion{
	"teacher": {
		{Label: "Check availability", Message: "Show me the availability schedule for this teacher.", Icon: "calendar"},
		{Label: "View schedule", Message: "What is this teacher's current timetable assignment?", Icon: "clock"},
		{Label: "Assign to subject", Message: "Which subjects can this teacher be assigned to?", Icon: "book"},
	},
	"subject": {
		{Label: "View prerequisites", Message: "List the prerequisites for this subject.", Icon: "git-branch"},
		{Label: "Find qualified teachers", Message: "Which teachers are qualified to teach this subject?", Icon: "users"},
	},
	"room": {
		{Label: "Check availability", Message: "When is this room available this semester?", Icon: "calendar"},
		{Label: "View assignments", Message: "What classes are assigned to this room?", Icon: "layout"},
	},
	"timetable": {
		{Label: "Explain conflicts", Message: "Are there any scheduling conflicts in this timetable?", Icon: "alert-triangle"},
		{Label: "Optimize schedule", Message: "Can the current schedule be optimized for fewer gaps?", Icon: "zap"},
	},
	"semester": {
		{Label: "View summary", Message: "Give me a summary of this semester's schedule.", Icon: "bar-chart"},
		{Label: "Check completion", Message: "How many subjects still need to be scheduled this semester?", Icon: "check-circle"},
	},
}

// HandleSuggestions handles GET /api/v1/agent/suggestions?entity_type=X&entity_id=Y
func (h *SuggestionHandler) HandleSuggestions(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	suggestions, ok := suggestionRules[entityType]
	if !ok {
		// Unknown entity type — return empty list, not an error.
		writeJSON(w, http.StatusOK, map[string]any{"suggestions": []Suggestion{}})
		return
	}

	// Enrich messages with entity_id so the chat handler has context.
	enriched := make([]Suggestion, len(suggestions))
	for i, s := range suggestions {
		enriched[i] = s
		if entityID != "" {
			enriched[i].Message = s.Message + " (entity_id: " + entityID + ")"
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entity_type": entityType,
		"entity_id":   entityID,
		"suggestions": enriched,
	})
}
