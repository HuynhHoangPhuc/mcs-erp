package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/scheduler"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// ProblemBuilder assembles a scheduler.Problem from cross-module data.
type ProblemBuilder interface {
	BuildProblem(ctx context.Context, semesterSubjects []*domain.SemesterSubject) (scheduler.Problem, error)
}

// ScheduleHandler handles schedule generation, retrieval, approval and manual edits.
type ScheduleHandler struct {
	semesterRepo  domain.SemesterRepository
	scheduleRepo  domain.ScheduleRepository
	problemBuilder ProblemBuilder
}

// NewScheduleHandler creates a new schedule handler.
func NewScheduleHandler(
	semesterRepo domain.SemesterRepository,
	scheduleRepo domain.ScheduleRepository,
	problemBuilder ProblemBuilder,
) *ScheduleHandler {
	return &ScheduleHandler{
		semesterRepo:   semesterRepo,
		scheduleRepo:   scheduleRepo,
		problemBuilder: problemBuilder,
	}
}

// GenerateSchedule handles POST /api/v1/timetable/semesters/{id}/generate
// Runs GreedyAssign + ParallelAnneal synchronously and persists the result.
func (h *ScheduleHandler) GenerateSchedule(w http.ResponseWriter, r *http.Request) {
	semID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid semester id"))
		return
	}

	sem, err := h.semesterRepo.FindByID(r.Context(), semID)
	if err != nil {
		if err == erptypes.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp("semester not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("failed to load semester"))
		return
	}

	// Mark semester as scheduling.
	sem.Status = domain.SemesterStatusScheduling
	if err := h.semesterRepo.Update(r.Context(), sem); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to update semester status"))
		return
	}

	// Load semester subjects.
	semSubjects, err := h.semesterRepo.GetSubjects(r.Context(), semID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to load semester subjects"))
		return
	}
	if len(semSubjects) == 0 {
		writeJSON(w, http.StatusBadRequest, errResp("semester has no subjects"))
		return
	}

	// Build the scheduling problem.
	problem, err := h.problemBuilder.BuildProblem(r.Context(), semSubjects)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to build scheduling problem"))
		return
	}

	// Run parallel simulated annealing.
	cfg := scheduler.DefaultSAConfig()
	assignments := scheduler.ParallelAnneal(problem, cfg, 0)

	// Stamp semester ID and compute metrics.
	hard := scheduler.BuildHardConstraints(problem)
	soft := scheduler.BuildSoftConstraints()

	for i := range assignments {
		assignments[i].SemesterID = semID
	}

	hardViolations := 0
	for _, c := range hard {
		hardViolations += c.Evaluate(assignments)
	}
	softPenalty := 0
	for _, c := range soft {
		softPenalty += c.Evaluate(assignments)
	}

	// Determine next version number.
	version := 1
	if latest, err := h.scheduleRepo.FindLatestBySemester(r.Context(), semID); err == nil {
		version = latest.Version + 1
	}

	// Tag each assignment with version.
	for i := range assignments {
		assignments[i].Version = version
		if assignments[i].ID == uuid.Nil {
			assignments[i].ID = uuid.New()
		}
	}

	sched := &domain.Schedule{
		SemesterID:     semID,
		Version:        version,
		Assignments:    assignments,
		HardViolations: hardViolations,
		SoftPenalty:    float64(softPenalty),
		GeneratedAt:    time.Now(),
	}

	if err := h.scheduleRepo.Save(r.Context(), sched); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to save schedule"))
		return
	}

	// Move semester to review status.
	sem.Status = domain.SemesterStatusReview
	_ = h.semesterRepo.Update(r.Context(), sem)

	writeJSON(w, http.StatusOK, scheduleResponse(sched))
}

// GetLatestSchedule handles GET /api/v1/timetable/semesters/{id}/schedule
func (h *ScheduleHandler) GetLatestSchedule(w http.ResponseWriter, r *http.Request) {
	semID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid semester id"))
		return
	}

	sched, err := h.scheduleRepo.FindLatestBySemester(r.Context(), semID)
	if err != nil {
		if err == erptypes.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp("no schedule found for semester"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("failed to load schedule"))
		return
	}
	writeJSON(w, http.StatusOK, scheduleResponse(sched))
}

// ApproveSchedule handles POST /api/v1/timetable/semesters/{id}/approve
func (h *ScheduleHandler) ApproveSchedule(w http.ResponseWriter, r *http.Request) {
	semID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid semester id"))
		return
	}

	sem, err := h.semesterRepo.FindByID(r.Context(), semID)
	if err != nil {
		if err == erptypes.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp("semester not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("failed to load semester"))
		return
	}

	if sem.Status != domain.SemesterStatusReview {
		writeJSON(w, http.StatusBadRequest, errResp("semester must be in review status to approve"))
		return
	}

	sem.Status = domain.SemesterStatusApproved
	if err := h.semesterRepo.Update(r.Context(), sem); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to approve schedule"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": sem.Status})
}

// UpdateAssignment handles PUT /api/v1/timetable/assignments/{id}
func (h *ScheduleHandler) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	assignID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid assignment id"))
		return
	}

	existing, err := h.scheduleRepo.FindAssignmentByID(r.Context(), assignID)
	if err != nil {
		if err == erptypes.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp("assignment not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("failed to load assignment"))
		return
	}

	var req struct {
		TeacherID string `json:"teacher_id"`
		RoomID    string `json:"room_id"`
		Day       int    `json:"day"`
		Period    int    `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}

	if req.TeacherID != "" {
		tid, err := uuid.Parse(req.TeacherID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResp("invalid teacher_id"))
			return
		}
		existing.TeacherID = tid
	}
	if req.RoomID != "" {
		rid, err := uuid.Parse(req.RoomID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResp("invalid room_id"))
			return
		}
		existing.RoomID = rid
	}
	if req.Day >= 0 && req.Day <= 5 {
		existing.Day = req.Day
	}
	if req.Period >= 1 && req.Period <= 10 {
		existing.Period = req.Period
	}

	if err := h.scheduleRepo.UpdateAssignment(r.Context(), existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to update assignment"))
		return
	}
	writeJSON(w, http.StatusOK, assignmentResponse(existing))
}

// --- Helpers ---

func scheduleResponse(s *domain.Schedule) map[string]any {
	assignments := make([]map[string]any, len(s.Assignments))
	for i, a := range s.Assignments {
		assignments[i] = assignmentResponse(&a)
	}
	return map[string]any{
		"semester_id":     s.SemesterID,
		"version":         s.Version,
		"hard_violations": s.HardViolations,
		"soft_penalty":    s.SoftPenalty,
		"generated_at":    s.GeneratedAt,
		"assignments":     assignments,
	}
}

func assignmentResponse(a *domain.Assignment) map[string]any {
	return map[string]any{
		"id":          a.ID,
		"semester_id": a.SemesterID,
		"subject_id":  a.SubjectID,
		"teacher_id":  a.TeacherID,
		"room_id":     a.RoomID,
		"day":         a.Day,
		"period":      a.Period,
		"version":     a.Version,
	}
}
