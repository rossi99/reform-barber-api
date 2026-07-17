package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/auth"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/pgconv"
	"github.com/reform-barber/api/internal/respond"
)

type AdminHandler struct {
	q *db.Queries
}

func NewAdminHandler(pool *pgxpool.Pool) *AdminHandler {
	return &AdminHandler{q: db.New(pool)}
}

// ListUsers returns every user in the system for admin management.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListUsers(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load users")
		return
	}
	respond.OK(w, rows)
}

type updateUserRoleRequest struct {
	Role string `json:"role"`
}

// UpdateUserRole promotes/demotes a user between customer, barber, founder, and admin.
// It refuses to demote the last remaining admin so the site can't be locked out.
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req updateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !auth.ValidRole(req.Role) {
		respond.Error(w, http.StatusBadRequest, "unknown role")
		return
	}

	target, err := h.q.GetUserByID(r.Context(), id)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}

	if target.Role == auth.RoleAdmin && req.Role != auth.RoleAdmin {
		count, err := h.q.CountUsersByRole(r.Context(), auth.RoleAdmin)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "could not verify admin count")
			return
		}
		if count <= 1 {
			respond.Error(w, http.StatusConflict, "cannot demote the last remaining admin")
			return
		}
	}

	updated, err := h.q.UpdateUserRole(r.Context(), db.UpdateUserRoleParams{ID: id, Role: req.Role})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not update role")
		return
	}
	respond.OK(w, map[string]any{
		"id":        updated.ID,
		"email":     updated.Email,
		"firstName": updated.FirstName,
		"lastName":  updated.LastName,
		"role":      updated.Role,
	})
}

// Stats returns site-wide booking counts and revenue for a date range,
// defaulting to the current calendar month.
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	from, to, err := statsDateRange(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "from/to must be YYYY-MM-DD")
		return
	}

	stats, err := h.q.BookingStats(r.Context(), db.BookingStatsParams{
		Date:   pgconv.Date(from),
		Date_2: pgconv.Date(to),
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load stats")
		return
	}

	respond.OK(w, map[string]any{
		"from":           from.Format("2006-01-02"),
		"to":             to.Format("2006-01-02"),
		"confirmedCount": stats.ConfirmedCount,
		"cancelledCount": stats.CancelledCount,
		"completedCount": stats.CompletedCount,
		"totalPence":     stats.TotalPence,
	})
}

func statsDateRange(r *http.Request) (time.Time, time.Time, error) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, -1)

	if v := r.URL.Query().Get("from"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed
	}
	if v := r.URL.Query().Get("to"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed
	}
	return from, to, nil
}
