package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/model"
	"github.com/reform-barber/api/internal/respond"
)

type ServicesHandler struct {
	q *db.Queries
}

func NewServicesHandler(pool *pgxpool.Pool) *ServicesHandler {
	return &ServicesHandler{q: db.New(pool)}
}

func (h *ServicesHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListServices(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load services")
		return
	}
	out := make([]model.Service, 0, len(rows))
	for _, s := range rows {
		out = append(out, model.Service{
			ID:           s.ID,
			Num:          derefStr(s.Num),
			Name:         s.Name,
			NameHTML:     derefStr(s.NameHtml),
			Description:  derefStr(s.Description),
			DurationMins: s.DurationMins,
			PricePence:   s.PricePence,
			Published:    s.Active,
		})
	}
	respond.OK(w, out)
}

func (h *ServicesHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListAllServices(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load services")
		return
	}
	out := make([]model.Service, 0, len(rows))
	for _, s := range rows {
		out = append(out, model.Service{
			ID:           s.ID,
			Num:          derefStr(s.Num),
			Name:         s.Name,
			NameHTML:     derefStr(s.NameHtml),
			Description:  derefStr(s.Description),
			DurationMins: s.DurationMins,
			PricePence:   s.PricePence,
			Published:    s.Active,
		})
	}
	respond.OK(w, out)
}

func (h *ServicesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Num          string `json:"num"`
		Name         string `json:"name"`
		NameHTML     string `json:"nameHtml"`
		Description  string `json:"description"`
		DurationMins int32  `json:"duration"`
		PricePence   int32  `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.DurationMins == 0 || body.PricePence == 0 {
		respond.Error(w, http.StatusBadRequest, "name, duration, and price are required")
		return
	}
	svc, err := h.q.CreateService(r.Context(), db.CreateServiceParams{
		Num:          &body.Num,
		Name:         body.Name,
		NameHtml:     &body.NameHTML,
		Description:  &body.Description,
		DurationMins: body.DurationMins,
		PricePence:   body.PricePence,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not create service")
		return
	}
	respond.Created(w, svc)
}

func (h *ServicesHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid service id")
		return
	}
	var body struct {
		Published bool `json:"published"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := h.q.SetServicePublished(r.Context(), db.SetServicePublishedParams{
		ID:     id,
		Active: body.Published,
	})
	if err != nil {
		respond.Error(w, http.StatusNotFound, "service not found")
		return
	}
	respond.OK(w, model.Service{
		ID:           svc.ID,
		Num:          derefStr(svc.Num),
		Name:         svc.Name,
		NameHTML:     derefStr(svc.NameHtml),
		Description:  derefStr(svc.Description),
		DurationMins: svc.DurationMins,
		PricePence:   svc.PricePence,
		Published:    svc.Active,
	})
}

func (h *ServicesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid service id")
		return
	}
	var body struct {
		Num          string `json:"num"`
		Name         string `json:"name"`
		NameHTML     string `json:"nameHtml"`
		Description  string `json:"description"`
		DurationMins int32  `json:"duration"`
		PricePence   int32  `json:"price"`
		Active       bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := h.q.UpdateService(r.Context(), db.UpdateServiceParams{
		ID:           id,
		Num:          &body.Num,
		Name:         body.Name,
		NameHtml:     &body.NameHTML,
		Description:  &body.Description,
		DurationMins: body.DurationMins,
		PricePence:   body.PricePence,
		Active:       body.Active,
	})
	if err != nil {
		respond.Error(w, http.StatusNotFound, "service not found")
		return
	}
	respond.OK(w, svc)
}
