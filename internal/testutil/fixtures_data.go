package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	timetabledomain "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

// SeedAdmin creates an admin user with full permissions and users_lookup entry.
func SeedAdmin(t *testing.T, pool *pgxpool.Pool, schema string) *SeedResult {
	t.Helper()

	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()
	email := fmt.Sprintf("admin_%s@example.com", uuid.NewString())
	password := defaultAdminPassword

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO public.tenants (name, schema_name, is_active, created_at, updated_at)
		 VALUES ($1, $2, true, now(), now())
		 ON CONFLICT (schema_name) DO NOTHING`,
		"Tenant "+schema, schema,
	); err != nil {
		t.Fatalf("insert public.tenants: %v", err)
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO public.users_lookup (email, tenant_schema)
		 VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET tenant_schema = EXCLUDED.tenant_schema`,
		email, schema,
	); err != nil {
		t.Fatalf("upsert public.users_lookup: %v", err)
	}

	err = database.WithTenantTx(ctx, pool, schema, func(tx pgx.Tx) error {
		now := time.Now().UTC()
		if _, err := tx.Exec(ctx,
			`INSERT INTO roles (id, name, permissions, description, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			roleID,
			"admin_"+uuid.NewString(),
			coredomain.AllPermissions(),
			"integration test admin role",
			now,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, name, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, true, $5, $5)`,
			userID,
			email,
			string(hashed),
			"Integration Admin",
			now,
		); err != nil {
			return err
		}

		_, err := tx.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
			userID, roleID,
		)
		return err
	})
	if err != nil {
		t.Fatalf("seed tenant admin: %v", err)
	}

	return &SeedResult{
		UserID:   userID,
		RoleID:   roleID,
		Email:    email,
		Password: password,
		Schema:   schema,
	}
}

// SeedTeacher creates a teacher and a default available slot.
func SeedTeacher(t *testing.T, pool *pgxpool.Pool, schema string, opts ...TeacherOption) *TeacherFixture {
	t.Helper()

	cfg := teacherSeedOpts{
		Name:  "Teacher " + uuid.NewString()[:8],
		Email: fmt.Sprintf("teacher_%s@example.com", uuid.NewString()),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	teacherID := uuid.New()
	deptID := cfg.DepartmentID

	err := database.WithTenantTx(context.Background(), pool, schema, func(tx pgx.Tx) error {
		if deptID == nil {
			id := uuid.New()
			deptID = &id
			if _, err := tx.Exec(context.Background(),
				`INSERT INTO departments (id, name, description, created_at)
				 VALUES ($1, $2, $3, now())`,
				id,
				"Department "+id.String()[:8],
				"integration test department",
			); err != nil {
				return err
			}
		}

		if _, err := tx.Exec(context.Background(),
			`INSERT INTO teachers (id, name, email, department_id, qualifications, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, true, now(), now())`,
			teacherID,
			cfg.Name,
			cfg.Email,
			deptID,
			[]string{"MSc"},
		); err != nil {
			return err
		}

		_, err := tx.Exec(context.Background(),
			`INSERT INTO teacher_availability (teacher_id, day, period, is_available)
			 VALUES ($1, 0, 1, true)
			 ON CONFLICT (teacher_id, day, period) DO UPDATE SET is_available = EXCLUDED.is_available`,
			teacherID,
		)
		return err
	})
	if err != nil {
		t.Fatalf("seed teacher: %v", err)
	}

	return &TeacherFixture{ID: teacherID, DepartmentID: deptID, Name: cfg.Name, Email: cfg.Email}
}

// SeedSubject creates a subject row in the tenant schema.
func SeedSubject(t *testing.T, pool *pgxpool.Pool, schema string, opts ...SubjectOption) *SubjectFixture {
	t.Helper()

	cfg := subjectSeedOpts{
		Name:         "Subject " + uuid.NewString()[:8],
		Code:         "SUB-" + stringsNoHyphen(uuid.NewString())[:8],
		Description:  "integration test subject",
		Credits:      3,
		HoursPerWeek: 3,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	subjectID := uuid.New()
	err := database.WithTenantTx(context.Background(), pool, schema, func(tx pgx.Tx) error {
		if cfg.CategoryID != nil {
			if _, err := tx.Exec(context.Background(),
				`INSERT INTO subject_categories (id, name, description, created_at)
				 VALUES ($1, $2, $3, now())
				 ON CONFLICT (id) DO NOTHING`,
				*cfg.CategoryID,
				"Category "+cfg.CategoryID.String()[:8],
				"integration test category",
			); err != nil {
				return err
			}
		}

		_, err := tx.Exec(context.Background(),
			`INSERT INTO subjects (id, name, code, description, category_id, credits, hours_per_week, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, true, now(), now())`,
			subjectID,
			cfg.Name,
			cfg.Code,
			cfg.Description,
			cfg.CategoryID,
			cfg.Credits,
			cfg.HoursPerWeek,
		)
		return err
	})
	if err != nil {
		t.Fatalf("seed subject: %v", err)
	}

	return &SubjectFixture{ID: subjectID, Name: cfg.Name, Code: cfg.Code}
}

// SeedRoom creates a room and a default available slot.
func SeedRoom(t *testing.T, pool *pgxpool.Pool, schema string, opts ...RoomOption) *RoomFixture {
	t.Helper()

	cfg := roomSeedOpts{
		Name:     "Room " + uuid.NewString()[:8],
		Code:     "R-" + stringsNoHyphen(uuid.NewString())[:6],
		Capacity: 40,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	roomID := uuid.New()
	err := database.WithTenantTx(context.Background(), pool, schema, func(tx pgx.Tx) error {
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO rooms (id, name, code, building, floor, capacity, equipment, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, true, now(), now())`,
			roomID,
			cfg.Name,
			cfg.Code,
			"A",
			1,
			cfg.Capacity,
			[]string{"projector"},
		); err != nil {
			return err
		}

		_, err := tx.Exec(context.Background(),
			`INSERT INTO room_availability (room_id, day, period, is_available)
			 VALUES ($1, 0, 1, true)
			 ON CONFLICT (room_id, day, period) DO UPDATE SET is_available = EXCLUDED.is_available`,
			roomID,
		)
		return err
	})
	if err != nil {
		t.Fatalf("seed room: %v", err)
	}

	return &RoomFixture{ID: roomID, Name: cfg.Name, Code: cfg.Code, Capacity: cfg.Capacity}
}

// SeedSemester creates a semester row in draft state by default.
func SeedSemester(t *testing.T, pool *pgxpool.Pool, schema string, opts ...SemesterOption) *SemesterFixture {
	t.Helper()

	start := time.Now().UTC().AddDate(0, 0, -1)
	end := start.AddDate(0, 4, 0)
	cfg := semesterSeedOpts{
		Name:      "Semester " + uuid.NewString()[:8],
		StartDate: start,
		EndDate:   end,
		Status:    timetabledomain.SemesterStatusDraft,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	semesterID := uuid.New()
	err := database.WithTenantTx(context.Background(), pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(),
			`INSERT INTO semesters (id, name, start_date, end_date, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, now(), now())`,
			semesterID,
			cfg.Name,
			cfg.StartDate,
			cfg.EndDate,
			string(cfg.Status),
		)
		return err
	})
	if err != nil {
		t.Fatalf("seed semester: %v", err)
	}

	return &SemesterFixture{
		ID:        semesterID,
		Name:      cfg.Name,
		StartDate: cfg.StartDate,
		EndDate:   cfg.EndDate,
		Status:    cfg.Status,
	}
}
