package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/middleware"
	"github.com/reform-barber/api/internal/model"
	"github.com/reform-barber/api/internal/pgconv"
	"github.com/reform-barber/api/internal/respond"
	"github.com/reform-barber/api/internal/storage"
)

const maxUploadBytes = 10 << 20 // 10 MB

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type MediaHandler struct {
	q     *db.Queries
	store storage.Store
}

func NewMediaHandler(pool *pgxpool.Pool, store storage.Store) *MediaHandler {
	return &MediaHandler{q: db.New(pool), store: store}
}

func (h *MediaHandler) ListCarousel(w http.ResponseWriter, r *http.Request) {
	h.listByType(w, r, "carousel")
}

func (h *MediaHandler) ListGallery(w http.ResponseWriter, r *http.Request) {
	h.listByType(w, r, "gallery")
}

func (h *MediaHandler) listByType(w http.ResponseWriter, r *http.Request, mediaType string) {
	rows, err := h.q.ListMediaByType(r.Context(), mediaType)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load media")
		return
	}
	out := make([]model.MediaItem, 0, len(rows))
	for _, m := range rows {
		item := model.MediaItem{
			ID:        m.ID,
			Type:      m.Type,
			PublicURL: m.PublicUrl,
			SortOrder: m.SortOrder,
		}
		if m.AltText != nil {
			item.AltText = *m.AltText
		}
		out = append(out, item)
	}
	respond.OK(w, out)
}

func (h *MediaHandler) UploadBarberPhoto(w http.ResponseWriter, r *http.Request) {
	barberID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid barber id")
		return
	}
	h.uploadEntity(w, r, "barber_photo", barberID)
}

func (h *MediaHandler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid product id")
		return
	}
	h.uploadEntity(w, r, "product_image", productID)
}

func (h *MediaHandler) UploadCarousel(w http.ResponseWriter, r *http.Request) {
	h.uploadFreeform(w, r, "carousel")
}

func (h *MediaHandler) UploadGallery(w http.ResponseWriter, r *http.Request) {
	h.uploadFreeform(w, r, "gallery")
}

func (h *MediaHandler) uploadEntity(w http.ResponseWriter, r *http.Request, mediaType string, entityID uuid.UUID) {
	key, publicURL, err := h.receiveUpload(w, r, mediaType)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	uploaderID := middleware.UserIDFromCtx(r.Context())

	_ = h.q.DeleteMediaForEntity(r.Context(), db.DeleteMediaForEntityParams{
		Type:     mediaType,
		EntityID: pgconv.UUID(entityID),
	})

	m, err := h.q.CreateMedia(r.Context(), db.CreateMediaParams{
		Type:       mediaType,
		EntityID:   pgconv.UUID(entityID),
		BucketKey:  key,
		PublicUrl:  publicURL,
		SortOrder:  0,
		UploadedBy: pgconv.UUID(uploaderID),
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not save media record")
		return
	}
	respond.Created(w, model.MediaItem{ID: m.ID, Type: m.Type, PublicURL: m.PublicUrl})
}

func (h *MediaHandler) uploadFreeform(w http.ResponseWriter, r *http.Request, mediaType string) {
	key, publicURL, err := h.receiveUpload(w, r, mediaType)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	altText := r.FormValue("alt")
	uploaderID := middleware.UserIDFromCtx(r.Context())

	m, err := h.q.CreateMedia(r.Context(), db.CreateMediaParams{
		Type:       mediaType,
		BucketKey:  key,
		PublicUrl:  publicURL,
		AltText:    &altText,
		UploadedBy: pgconv.UUID(uploaderID),
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not save media record")
		return
	}
	respond.Created(w, model.MediaItem{ID: m.ID, Type: m.Type, PublicURL: m.PublicUrl})
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid media id")
		return
	}

	m, err := h.q.GetMediaByID(r.Context(), id)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "media not found")
		return
	}

	_ = h.store.Delete(r.Context(), m.BucketKey)
	_ = h.q.DeleteMedia(r.Context(), id)
	respond.NoContent(w)
}

func (h *MediaHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid media id")
		return
	}
	var body struct {
		SortOrder int32 `json:"sortOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_ = h.q.UpdateMediaSortOrder(r.Context(), db.UpdateMediaSortOrderParams{ID: id, SortOrder: body.SortOrder})
	respond.NoContent(w)
}

func (h *MediaHandler) receiveUpload(w http.ResponseWriter, r *http.Request, mediaType string) (key, publicURL string, err error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err = r.ParseMultipartForm(maxUploadBytes); err != nil {
		return "", "", fmt.Errorf("file too large or bad form data")
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return "", "", fmt.Errorf("image field is required")
	}
	defer file.Close()

	contentType := strings.Split(header.Header.Get("Content-Type"), ";")[0]
	ext, ok := allowedMIME[contentType]
	if !ok {
		return "", "", fmt.Errorf("unsupported image type: %s (allowed: jpeg, png, webp)", contentType)
	}

	key = fmt.Sprintf("%s/%s%s", mediaType, uuid.New().String(), ext)
	publicURL, err = h.store.Upload(r.Context(), key, file, contentType)
	if err != nil {
		return "", "", fmt.Errorf("upload failed: %w", err)
	}
	return key, publicURL, nil
}
