package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// AvailabilityHandler handles room availability endpoints.
type AvailabilityHandler struct {
	roomRepo  domain.RoomRepository
	availRepo domain.RoomAvailabilityRepository
}

// NewAvailabilityHandler creates a new AvailabilityHandler.
func NewAvailabilityHandler(roomRepo domain.RoomRepository, availRepo domain.RoomAvailabilityRepository) *AvailabilityHandler {
	return &AvailabilityHandler{roomRepo: roomRepo, availRepo: availRepo}
}

// slotRequest is a single slot entry in the PUT body.
type slotRequest struct {
	Day         int  `json:"day"`    // 0-6
	Period      int  `json:"period"` // 1-10
	IsAvailable bool `json:"is_available"`
}

// slotResponse is a single slot entry in the GET response.
type slotResponse struct {
	Day         int  `json:"day"`
	Period      int  `json:"period"`
	IsAvailable bool `json:"is_available"`
}

// GetAvailability handles GET /api/v1/rooms/{id}/availability
func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid room id"})
		return
	}

	// Verify the room exists
	if _, err := h.roomRepo.FindByID(r.Context(), roomID); err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get room"})
		return
	}

	avail, err := h.availRepo.GetByRoomID(r.Context(), roomID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get availability"})
		return
	}

	slots := make([]slotResponse, 0, len(avail))
	for slot, isAvailable := range avail {
		slots = append(slots, slotResponse{
			Day:         slot.Day,
			Period:      slot.Period,
			IsAvailable: isAvailable,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"room_id": roomID, "slots": slots})
}

// SetAvailability handles PUT /api/v1/rooms/{id}/availability
func (h *AvailabilityHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid room id"})
		return
	}

	// Verify the room exists
	if _, err := h.roomRepo.FindByID(r.Context(), roomID); err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get room"})
		return
	}

	var body struct {
		Slots []slotRequest `json:"slots"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate and build the availability map
	avail := make(domain.RoomAvailability, len(body.Slots))
	for _, s := range body.Slots {
		if s.Day < 0 || s.Day > 6 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "day must be between 0 and 6"})
			return
		}
		if s.Period < 1 || s.Period > 10 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "period must be between 1 and 10"})
			return
		}
		avail[domain.WeeklySlot{Day: s.Day, Period: s.Period}] = s.IsAvailable
	}

	if err := h.availRepo.SetSlots(r.Context(), roomID, avail); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to set availability"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "availability updated"})
}
