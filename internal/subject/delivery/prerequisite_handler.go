package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PrerequisiteHandler manages prerequisite edges between subjects.
// Cycle detection is performed in-memory using the DAG graph before persisting.
type PrerequisiteHandler struct {
	prereqRepo  domain.PrerequisiteRepository
	subjectRepo domain.SubjectRepository
}

// NewPrerequisiteHandler creates a new PrerequisiteHandler.
func NewPrerequisiteHandler(prereqRepo domain.PrerequisiteRepository, subjectRepo domain.SubjectRepository) *PrerequisiteHandler {
	return &PrerequisiteHandler{prereqRepo: prereqRepo, subjectRepo: subjectRepo}
}

type addPrerequisiteRequest struct {
	PrerequisiteID uuid.UUID `json:"prerequisite_id"`
	// ExpectedVersion is the caller's known version for optimistic locking.
	// If omitted the server will read the current version and use it
	// (last-write-wins per request, still safe within a single request).
	ExpectedVersion *int `json:"expected_version"`
}

// AddPrerequisite handles POST /api/v1/subjects/{id}/prerequisites
// Cycle detection steps:
//  1. Load all existing edges from DB
//  2. Build in-memory graph
//  3. Check if new edge creates a cycle
//  4. If safe, persist with optimistic locking
func (h *PrerequisiteHandler) AddPrerequisite(w http.ResponseWriter, r *http.Request) {
	subjectID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject id"})
		return
	}

	var req addPrerequisiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.PrerequisiteID == uuid.Nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prerequisite_id is required"})
		return
	}

	if subjectID == req.PrerequisiteID {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a subject cannot be its own prerequisite"})
		return
	}

	// Verify both subjects exist.
	if _, err := h.subjectRepo.FindByID(r.Context(), subjectID); err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subject not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to verify subject"})
		return
	}
	if _, err := h.subjectRepo.FindByID(r.Context(), req.PrerequisiteID); err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "prerequisite subject not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to verify prerequisite subject"})
		return
	}

	// Load all edges and build in-memory graph for cycle detection.
	allEdges, err := h.prereqRepo.GetAllEdges(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load prerequisite graph"})
		return
	}

	graph := domain.NewGraph(allEdges)
	if domain.HasCycle(graph, subjectID, req.PrerequisiteID) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "adding this prerequisite would create a cycle"})
		return
	}

	// Determine expected version for optimistic locking.
	expectedVersion := 0
	if req.ExpectedVersion != nil {
		expectedVersion = *req.ExpectedVersion
	} else {
		expectedVersion, err = h.prereqRepo.GetVersion(r.Context(), subjectID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get version"})
			return
		}
	}

	if err := h.prereqRepo.AddEdge(r.Context(), subjectID, req.PrerequisiteID, expectedVersion); err != nil {
		if errors.Is(err, erptypes.ErrConflict) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "concurrent modification detected, please retry"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add prerequisite"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"subject_id":      subjectID,
		"prerequisite_id": req.PrerequisiteID,
	})
}

// RemovePrerequisite handles DELETE /api/v1/subjects/{id}/prerequisites/{prereqId}
func (h *PrerequisiteHandler) RemovePrerequisite(w http.ResponseWriter, r *http.Request) {
	subjectID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject id"})
		return
	}

	prereqID, err := uuid.Parse(r.PathValue("prereqId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid prerequisite id"})
		return
	}

	// Resolve expected version from query param or DB.
	expectedVersion := 0
	if v := r.URL.Query().Get("expected_version"); v != "" {
		if _, err := parseIntParam(v, &expectedVersion); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid expected_version"})
			return
		}
	} else {
		expectedVersion, err = h.prereqRepo.GetVersion(r.Context(), subjectID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get version"})
			return
		}
	}

	if err := h.prereqRepo.RemoveEdge(r.Context(), subjectID, prereqID, expectedVersion); err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "prerequisite edge not found"})
			return
		}
		if errors.Is(err, erptypes.ErrConflict) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "concurrent modification detected, please retry"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove prerequisite"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "prerequisite removed"})
}

// ListPrerequisites handles GET /api/v1/subjects/{id}/prerequisites
func (h *PrerequisiteHandler) ListPrerequisites(w http.ResponseWriter, r *http.Request) {
	subjectID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject id"})
		return
	}

	edges, err := h.prereqRepo.GetEdges(r.Context(), subjectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list prerequisites"})
		return
	}

	items := make([]map[string]any, len(edges))
	for i, e := range edges {
		items[i] = map[string]any{
			"subject_id":      e.SubjectID,
			"prerequisite_id": e.PrerequisiteID,
		}
	}

	version := 0
	if len(edges) > 0 {
		version = edges[0].Version
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":   items,
		"version": version,
	})
}

// GetPrerequisiteChain handles GET /api/v1/subjects/{id}/prerequisite-chain
// Returns all transitive prerequisites (full DAG reachability).
func (h *PrerequisiteHandler) GetPrerequisiteChain(w http.ResponseWriter, r *http.Request) {
	subjectID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject id"})
		return
	}

	allEdges, err := h.prereqRepo.GetAllEdges(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load prerequisite graph"})
		return
	}

	graph := domain.NewGraph(allEdges)
	chain := domain.GetPrerequisiteChain(graph, subjectID)

	ids := make([]uuid.UUID, len(chain))
	copy(ids, chain)

	writeJSON(w, http.StatusOK, map[string]any{
		"subject_id": subjectID,
		"chain":      ids,
	})
}

// parseIntParam parses a string into an int pointer target. Returns error on failure.
func parseIntParam(s string, target *int) (int, error) {
	var v int
	if _, err := scanInt(s, &v); err != nil {
		return 0, err
	}
	*target = v
	return v, nil
}

func scanInt(s string, out *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("not an integer")
		}
		n = n*10 + int(c-'0')
	}
	*out = n
	return n, nil
}
