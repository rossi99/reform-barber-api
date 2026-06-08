package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/auth"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/middleware"
	"github.com/reform-barber/api/internal/pgconv"
	"github.com/reform-barber/api/internal/respond"
)

type AuthHandler struct {
	q         *db.Queries
	jwtSecret string
}

func NewAuthHandler(pool *pgxpool.Pool, jwtSecret string) *AuthHandler {
	return &AuthHandler{q: db.New(pool), jwtSecret: jwtSecret}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Phone       string `json:"phone"`
	ReminderOpt bool   `json:"reminderOpt"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respond.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < 8 {
		respond.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user, err := h.q.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
		FirstName:    &req.FirstName,
		LastName:     &req.LastName,
		Phone:        &req.Phone,
		Role:         "customer",
		ReminderOpt:  req.ReminderOpt,
	})
	if err != nil {
		respond.Error(w, http.StatusConflict, "email already registered")
		return
	}

	resp, err := h.issueTokens(r, user.ID, user.Role, nil)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not issue tokens")
		return
	}
	respond.Created(w, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.q.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	var barberID *uuid.UUID
	if user.Role == "barber" || user.Role == "founder" {
		b, err := h.q.GetBarberByUserID(r.Context(), user.ID)
		if err == nil {
			barberID = &b.ID
		}
	}

	resp, err := h.issueTokens(r, user.ID, user.Role, barberID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not issue tokens")
		return
	}
	respond.OK(w, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		respond.Error(w, http.StatusBadRequest, "refreshToken required")
		return
	}

	hash := auth.HashRefreshToken(body.RefreshToken)
	stored, err := h.q.GetRefreshToken(r.Context(), hash)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	_ = h.q.DeleteRefreshToken(r.Context(), hash)

	user, err := h.q.GetUserByID(r.Context(), stored.UserID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "user not found")
		return
	}

	var barberID *uuid.UUID
	if user.Role == "barber" || user.Role == "founder" {
		b, err := h.q.GetBarberByUserID(r.Context(), user.ID)
		if err == nil {
			barberID = &b.ID
		}
	}

	resp, err := h.issueTokens(r, user.ID, user.Role, barberID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not issue tokens")
		return
	}
	respond.OK(w, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.RefreshToken != "" {
		hash := auth.HashRefreshToken(body.RefreshToken)
		_ = h.q.DeleteRefreshToken(r.Context(), hash)
	}
	respond.NoContent(w)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	user, err := h.q.GetUserByID(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}
	respond.OK(w, map[string]any{
		"id":        user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"role":      user.Role,
	})
}

func (h *AuthHandler) issueTokens(r *http.Request, userID uuid.UUID, role string, barberID *uuid.UUID) (tokenResponse, error) {
	accessToken, err := auth.IssueAccessToken(h.jwtSecret, userID, role, barberID)
	if err != nil {
		return tokenResponse{}, err
	}

	raw, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return tokenResponse{}, err
	}

	_, err = h.q.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: pgconv.Timestamp(time.Now().Add(auth.RefreshTokenTTL)),
	})
	if err != nil {
		return tokenResponse{}, err
	}

	return tokenResponse{AccessToken: accessToken, RefreshToken: raw}, nil
}
