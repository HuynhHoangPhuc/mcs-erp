package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
)

// AvailabilityHandler handles teacher availability endpoints.
type AvailabilityHandler struct {
	availRepo  domain.AvailabilityRepository
	teacherRepo domain.TeacherRepository
}

// NewAvailabilityHandler creates a new availability handler.
func NewAvailabilityHandler(availRepo domain.AvailabilityRepository, teacherRepo domain.TeacherRepository) *AvailabilityHandler {
	return &AvailabilityHandler{availRepo: availRepo, teacherRepo: teacherRepo}
}

// slotRequest represents a single day+period slot in the request body.
type slotRequest struct {
	Day         int  `json:"day"`    // 0-6
	Period      int  `json:"period"` // 1-10
	IsAvailable bool `json:"is_available"`
}

// setAvailabilityRequest is the payload for replacing teacher availability.
type setAvailabilityRequest struct {
	Slots []slotRequest `json:"slots"`
}

// GetAvailability handles GET /api/v1/teachers/{id}/availability
func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	teacherID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid teacher id"})
		return
	}

	// Verify teacher exists
	if _, err := h.teacherRepo.FindByID(r.Context(), teacherID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "teacher not found"})
		return
	}

	slots, err := h.availRepo.GetByTeacherID(r.Context(), teacherID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get availability"})
		return
	}

	items := make([]map[string]any, len(slots))
	for i, s := range slots {
		items[i] = map[string]any{
			"day":          s.Day,
			"period":       s.Period,
			"is_available": s.IsAvailable,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"teacher_id": teacherID,
		"slots":      items,
	})
}

// SetAvailability handles PUT /api/v1/teachers/{id}/availability
// Replaces all availability slots for the teacher.
func (h *AvailabilityHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	teacherID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid teacher id"})
		return
	}

	// Verify teacher exists
	if _, err := h.teacherRepo.FindByID(r.Context(), teacherID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "teacher not found"})
		return
	}

	var req setAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate slot values
	for _, s := range req.Slots {
		if s.Day < 0 || s.Day > 6 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "day must be 0-6"})
			return
		}
		if s.Period < 1 || s.Period > 10 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "period must be 1-10"})
			return
		}
	}

	slots := make([]*domain.Availability, len(req.Slots))
	for i, s := range req.Slots {
		slots[i] = &domain.Availability{
			TeacherID:   teacherID,
			Day:         s.Day,
			Period:      s.Period,
			IsAvailable: s.IsAvailable,
		}
	}

	if err := h.availRepo.SetSlots(r.Context(), teacherID, slots); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to set availability"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"teacher_id":  teacherID,
		"slots_count": len(slots),
		"message":     "availability updated",
	})
}
