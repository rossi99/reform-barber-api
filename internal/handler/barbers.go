package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/model"
	"github.com/reform-barber/api/internal/pgconv"
	"github.com/reform-barber/api/internal/respond"
)

type BarbersHandler struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewBarbersHandler(pool *pgxpool.Pool) *BarbersHandler {
	return &BarbersHandler{q: db.New(pool), pool: pool}
}

func (h *BarbersHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListBarbers(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load barbers")
		return
	}

	photos := map[uuid.UUID]string{}
	mediaRows, _ := h.q.ListMediaByType(r.Context(), "barber_photo")
	for _, m := range mediaRows {
		if eid, ok := pgconv.FromUUID(m.EntityID); ok {
			photos[eid] = m.PublicUrl
		}
	}

	out := make([]model.Barber, 0, len(rows))
	for _, b := range rows {
		bar := model.Barber{
			ID:    b.ID,
			Name:  b.Name,
			Num:   derefStr(b.Num),
			Title: derefStr(b.Title),
			Bio:   derefStr(b.Bio),
		}
		if url, ok := photos[b.ID]; ok {
			bar.PhotoURL = url
		}
		out = append(out, bar)
	}
	respond.OK(w, out)
}

func (h *BarbersHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid barber id")
		return
	}

	var body struct {
		Name   string `json:"name"`
		Title  string `json:"title"`
		Bio    string `json:"bio"`
		Num    string `json:"num"`
		Active bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	barber, err := h.q.UpdateBarber(r.Context(), db.UpdateBarberParams{
		ID:     id,
		Name:   body.Name,
		Title:  &body.Title,
		Bio:    &body.Bio,
		Num:    &body.Num,
		Active: body.Active,
	})
	if err != nil {
		respond.Error(w, http.StatusNotFound, "barber not found")
		return
	}
	respond.OK(w, barber)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
