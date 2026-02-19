package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// RoomHandler handles room CRUD endpoints.
type RoomHandler struct {
	roomRepo domain.RoomRepository
}

// NewRoomHandler creates a new RoomHandler.
func NewRoomHandler(roomRepo domain.RoomRepository) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo}
}

type createRoomRequest struct {
	Name      string   `json:"name"`
	Code      string   `json:"code"`
	Building  string   `json:"building"`
	Floor     int      `json:"floor"`
	Capacity  int      `json:"capacity"`
	Equipment []string `json:"equipment"`
}

type updateRoomRequest struct {
	Name      string   `json:"name"`
	Code      string   `json:"code"`
	Building  string   `json:"building"`
	Floor     int      `json:"floor"`
	Capacity  int      `json:"capacity"`
	Equipment []string `json:"equipment"`
	IsActive  *bool    `json:"is_active"`
}

// roomResponse is the standard JSON representation of a Room.
type roomResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Building  string    `json:"building"`
	Floor     int       `json:"floor"`
	Capacity  int       `json:"capacity"`
	Equipment []string  `json:"equipment"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toRoomResponse(r *domain.Room) roomResponse {
	eq := r.Equipment
	if eq == nil {
		eq = []string{}
	}
	return roomResponse{
		ID:        r.ID,
		Name:      r.Name,
		Code:      r.Code,
		Building:  r.Building,
		Floor:     r.Floor,
		Capacity:  r.Capacity,
		Equipment: eq,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// CreateRoom handles POST /api/v1/rooms
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and code are required"})
		return
	}
	if req.Capacity <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "capacity must be positive"})
		return
	}

	now := time.Now()
	room := &domain.Room{
		ID:        uuid.New(),
		Name:      req.Name,
		Code:      req.Code,
		Building:  req.Building,
		Floor:     req.Floor,
		Capacity:  req.Capacity,
		Equipment: req.Equipment,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.roomRepo.Save(r.Context(), room); err != nil {
		if errors.Is(err, erptypes.ErrConflict) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "room code already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create room"})
		return
	}

	writeJSON(w, http.StatusCreated, toRoomResponse(room))
}

// ListRooms handles GET /api/v1/rooms
// Query params: building, min_capacity, equipment (comma-separated)
func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := domain.ListFilter{
		Building: q.Get("building"),
	}

	if mc := q.Get("min_capacity"); mc != "" {
		n, err := strconv.Atoi(mc)
		if err != nil || n < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "min_capacity must be a non-negative integer"})
			return
		}
		filter.MinCapacity = n
	}

	if eq := q.Get("equipment"); eq != "" {
		for _, item := range strings.Split(eq, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				filter.Equipment = append(filter.Equipment, item)
			}
		}
	}

	rooms, err := h.roomRepo.List(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list rooms"})
		return
	}

	items := make([]roomResponse, len(rooms))
	for i, room := range rooms {
		items[i] = toRoomResponse(room)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

// GetRoom handles GET /api/v1/rooms/{id}
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid room id"})
		return
	}

	room, err := h.roomRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get room"})
		return
	}

	writeJSON(w, http.StatusOK, toRoomResponse(room))
}

// UpdateRoom handles PUT /api/v1/rooms/{id}
func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid room id"})
		return
	}

	var req updateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and code are required"})
		return
	}
	if req.Capacity <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "capacity must be positive"})
		return
	}

	room, err := h.roomRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "room not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get room"})
		return
	}

	room.Name = req.Name
	room.Code = req.Code
	room.Building = req.Building
	room.Floor = req.Floor
	room.Capacity = req.Capacity
	room.Equipment = req.Equipment
	if req.IsActive != nil {
		room.IsActive = *req.IsActive
	}

	if err := h.roomRepo.Update(r.Context(), room); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update room"})
		return
	}

	writeJSON(w, http.StatusOK, toRoomResponse(room))
}
