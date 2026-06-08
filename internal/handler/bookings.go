package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/middleware"
	"github.com/reform-barber/api/internal/model"
	"github.com/reform-barber/api/internal/notify"
	"github.com/reform-barber/api/internal/pgconv"
	"github.com/reform-barber/api/internal/respond"
)

type BookingsHandler struct {
	q      *db.Queries
	pool   *pgxpool.Pool
	notify notify.Notifier
}

func NewBookingsHandler(pool *pgxpool.Pool, n notify.Notifier) *BookingsHandler {
	return &BookingsHandler{q: db.New(pool), pool: pool, notify: n}
}

func (h *BookingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}

	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid service id")
		return
	}

	svc, err := h.q.GetServiceByID(r.Context(), serviceID)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "service not found")
		return
	}

	var barberID uuid.UUID
	if req.BarberID == "any" || req.BarberID == "" {
		barberID, err = h.firstAvailableBarber(r.Context(), date, req.Time, int(svc.DurationMins))
		if err != nil {
			respond.Error(w, http.StatusConflict, "no available barber for requested slot")
			return
		}
	} else {
		barberID, err = uuid.Parse(req.BarberID)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid barber id")
			return
		}
		existing, _ := h.q.ListConfirmedBookingsForBarberDate(r.Context(), db.ListConfirmedBookingsForBarberDateParams{
			BarberID: barberID,
			Date:     pgconv.Date(date),
		})
		startMins := timeStrToMins(req.Time)
		endMins := startMins + int(svc.DurationMins)
		if !isSlotFree(startEnd{startMins, endMins}, existing) {
			respond.Error(w, http.StatusConflict, "slot is no longer available")
			return
		}
	}

	startMins := timeStrToMins(req.Time)
	endMins := startMins + int(svc.DurationMins)
	reference := generateReference()
	userID := middleware.UserIDFromCtx(r.Context())

	booking, err := h.q.CreateBooking(r.Context(), db.CreateBookingParams{
		Reference: reference,
		UserID:    pgconv.UUID(userID),
		BarberID:  barberID,
		ServiceID: serviceID,
		Date:      pgconv.Date(date),
		TimeStart: req.Time,
		TimeEnd:   minsToTimeStr(endMins),
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not create booking")
		return
	}

	for productIDStr, qty := range req.Products {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			continue
		}
		p, err := h.q.GetProductByID(r.Context(), productID)
		if err != nil {
			continue
		}
		_, _ = h.q.CreateBookingProduct(r.Context(), db.CreateBookingProductParams{
			BookingID:  booking.ID,
			ProductID:  productID,
			Qty:        int32(qty),
			PricePence: p.PricePence,
		})
	}

	go func() {
		user, _ := h.q.GetUserByID(context.Background(), userID)
		barber, _ := h.q.GetBarberByID(context.Background(), barberID)
		payload := notify.BookingPayload{
			Reference:     reference,
			CustomerName:  fmt.Sprintf("%s %s", derefStr(user.FirstName), derefStr(user.LastName)),
			CustomerEmail: user.Email,
			CustomerPhone: derefStr(user.Phone),
			BarberName:    barber.Name,
			ServiceName:   svc.Name,
			Date:          date.Format("Mon, 2 Jan 2006"),
			Time:          req.Time,
		}
		_ = h.notify.BookingConfirmed(context.Background(), payload)
	}()

	respond.Created(w, model.BookingResponse{
		ID:        booking.ID,
		Reference: booking.Reference,
		Status:    booking.Status,
		BarberID:  booking.BarberID,
		ServiceID: booking.ServiceID,
		Date:      pgconv.FromDate(booking.Date).Format("2006-01-02"),
		Time:      booking.TimeStart,
		CreatedAt: pgconv.FromTimestamp(booking.CreatedAt),
	})
}

func (h *BookingsHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	rows, err := h.q.ListBookingsForUser(r.Context(), pgconv.UUID(userID))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load bookings")
		return
	}
	respond.OK(w, rows)
}

func (h *BookingsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid booking id")
		return
	}
	userID := middleware.UserIDFromCtx(r.Context())

	booking, err := h.q.GetBookingByID(r.Context(), id)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "booking not found")
		return
	}

	claims := middleware.ClaimsFromCtx(r.Context())
	ownerID, valid := pgconv.FromUUID(booking.UserID)
	if (!valid || ownerID != userID) && claims.Role != "founder" {
		respond.Error(w, http.StatusForbidden, "not your booking")
		return
	}

	updated, err := h.q.CancelBooking(r.Context(), id)
	if err != nil {
		respond.Error(w, http.StatusConflict, "booking cannot be cancelled")
		return
	}
	respond.OK(w, updated)
}

func (h *BookingsHandler) BarberAppointments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	if claims.BarberID == nil {
		respond.Error(w, http.StatusForbidden, "no barber profile linked to this account")
		return
	}

	dateParam := r.URL.Query().Get("date")
	if dateParam == "" {
		dateParam = time.Now().Format("2006-01-02")
	}
	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}

	rows, err := h.q.ListBookingsForBarber(r.Context(), db.ListBookingsForBarberParams{
		BarberID: *claims.BarberID,
		Date:     pgconv.Date(date),
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load appointments")
		return
	}
	respond.OK(w, rows)
}

func (h *BookingsHandler) AdminListBookings(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	barberParam := r.URL.Query().Get("barber")
	dateParam := r.URL.Query().Get("date")

	params := db.ListAllBookingsParams{}
	if statusFilter != "" {
		params.Status = &statusFilter
	}
	if barberParam != "" {
		bid, err := uuid.Parse(barberParam)
		if err == nil {
			params.BarberID = pgconv.UUID(bid)
		}
	}
	if dateParam != "" {
		d, err := time.Parse("2006-01-02", dateParam)
		if err == nil {
			params.Date = pgconv.Date(d)
		}
	}

	rows, err := h.q.ListAllBookings(r.Context(), params)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "could not load bookings")
		return
	}
	respond.OK(w, rows)
}

func (h *BookingsHandler) firstAvailableBarber(ctx context.Context, date time.Time, timeStr string, durationMins int) (uuid.UUID, error) {
	barberIDs, err := h.q.ListActiveBarberIDs(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	startMins := timeStrToMins(timeStr)
	endMins := startMins + durationMins

	for _, barberID := range barberIDs {
		existing, _ := h.q.ListConfirmedBookingsForBarberDate(ctx, db.ListConfirmedBookingsForBarberDateParams{
			BarberID: barberID,
			Date:     pgconv.Date(date),
		})
		if isSlotFree(startEnd{startMins, endMins}, existing) {
			return barberID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("no available barber")
}

func generateReference() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return "RFB-" + string(b)
}
