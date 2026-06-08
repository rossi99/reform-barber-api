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

type ProductsHandler struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewProductsHandler(pool *pgxpool.Pool) *ProductsHandler {
	return &ProductsHandler{q: db.New(pool), pool: pool}
}

func (h *ProductsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListProducts(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load products")
		return
	}

	photos := map[uuid.UUID]string{}
	mediaRows, _ := h.q.ListMediaByType(r.Context(), "product_image")
	for _, m := range mediaRows {
		if eid, ok := pgconv.FromUUID(m.EntityID); ok {
			photos[eid] = m.PublicUrl
		}
	}

	out := make([]model.Product, 0, len(rows))
	for _, p := range rows {
		prod := model.Product{
			ID:         p.ID,
			Name:       p.Name,
			PricePence: p.PricePence,
		}
		if url, ok := photos[p.ID]; ok {
			prod.ImageURL = url
		}
		out = append(out, prod)
	}
	respond.OK(w, out)
}

func (h *ProductsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		PricePence int32  `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.PricePence == 0 {
		respond.Error(w, http.StatusBadRequest, "name and price are required")
		return
	}
	p, err := h.q.CreateProduct(r.Context(), db.CreateProductParams{Name: body.Name, PricePence: body.PricePence})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not create product")
		return
	}
	respond.Created(w, p)
}

func (h *ProductsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var body struct {
		Name       string `json:"name"`
		PricePence int32  `json:"price"`
		Active     bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.q.UpdateProduct(r.Context(), db.UpdateProductParams{
		ID:         id,
		Name:       body.Name,
		PricePence: body.PricePence,
		Active:     body.Active,
	})
	if err != nil {
		respond.Error(w, http.StatusNotFound, "product not found")
		return
	}
	respond.OK(w, p)
}
