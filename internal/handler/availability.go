package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reform-barber/api/internal/db"
	"github.com/reform-barber/api/internal/model"
	"github.com/reform-barber/api/internal/pgconv"
	"github.com/reform-barber/api/internal/respond"
)

func GetAvailability(pool *pgxpool.Pool) http.HandlerFunc {
	q := db.New(pool)
	return func(w http.ResponseWriter, r *http.Request) {
		barberParam := r.URL.Query().Get("barber")
		serviceParam := r.URL.Query().Get("service")
		dateParam := r.URL.Query().Get("date")

		if dateParam == "" || serviceParam == "" {
			respond.Error(w, http.StatusBadRequest, "service and date are required")
			return
		}

		date, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
			return
		}

		serviceID, err := uuid.Parse(serviceParam)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid service id")
			return
		}

		svc, err := q.GetServiceByID(r.Context(), serviceID)
		if err != nil {
			respond.Error(w, http.StatusNotFound, "service not found")
			return
		}

		var barberIDs []uuid.UUID
		if barberParam == "any" || barberParam == "" {
			ids, err := q.ListActiveBarberIDs(r.Context())
			if err != nil {
				respond.Error(w, http.StatusInternalServerError, "could not load barbers")
				return
			}
			barberIDs = ids
		} else {
			bid, err := uuid.Parse(barberParam)
			if err != nil {
				respond.Error(w, http.StatusBadRequest, "invalid barber id")
				return
			}
			barberIDs = []uuid.UUID{bid}
		}

		// slotBooked: true = all barbers booked, false = at least one free.
		// Initialized to true; set false when any barber has it free.
		slotBooked := map[string]bool{}

		for _, barberID := range barberIDs {
			schedule, err := q.GetScheduleForBarberDOW(r.Context(), db.GetScheduleForBarberDOWParams{
				BarberID: barberID,
				Dow:      int32(date.Weekday()),
			})
			if err != nil {
				continue // not scheduled this day
			}

			openMins := timeStrToMins(schedule.OpenTime)
			closeMins := timeStrToMins(schedule.CloseTime)
			dur := int(svc.DurationMins)

			existing, _ := q.ListConfirmedBookingsForBarberDate(r.Context(), db.ListConfirmedBookingsForBarberDateParams{
				BarberID: barberID,
				Date:     pgconv.Date(date),
			})

			for m := openMins; m+dur <= closeMins; m += 10 {
				t := minsToTimeStr(m)
				if already, seen := slotBooked[t]; seen && !already {
					continue // already found a free barber for this slot
				}
				if isSlotFree(startEnd{m, m + dur}, existing) {
					slotBooked[t] = false // at least one barber is free
				} else if _, seen := slotBooked[t]; !seen {
					slotBooked[t] = true // no free barber yet
				}
			}
		}

		slots := make([]model.Slot, 0, len(slotBooked))
		for t, booked := range slotBooked {
			slots = append(slots, model.Slot{Time: t, Booked: booked})
		}
		sortSlots(slots)

		resp := model.AvailabilityResponse{Date: dateParam}
		for _, s := range slots {
			mins := timeStrToMins(s.Time)
			switch {
			case mins < 12*60:
				resp.Groups.Morning = append(resp.Groups.Morning, s)
			case mins < 17*60:
				resp.Groups.Afternoon = append(resp.Groups.Afternoon, s)
			default:
				resp.Groups.Evening = append(resp.Groups.Evening, s)
			}
		}
		if resp.Groups.Morning == nil {
			resp.Groups.Morning = []model.Slot{}
		}
		if resp.Groups.Afternoon == nil {
			resp.Groups.Afternoon = []model.Slot{}
		}
		if resp.Groups.Evening == nil {
			resp.Groups.Evening = []model.Slot{}
		}

		respond.OK(w, resp)
	}
}

type startEnd struct{ start, end int }

func isSlotFree(slot startEnd, booked []db.ListConfirmedBookingsForBarberDateRow) bool {
	for _, b := range booked {
		bStart := timeStrToMins(b.TimeStart)
		bEnd := timeStrToMins(b.TimeEnd)
		if slot.start < bEnd && slot.end > bStart {
			return false
		}
	}
	return true
}

func timeStrToMins(t string) int {
	var h, m int
	fmt.Sscanf(t, "%d:%d", &h, &m)
	return h*60 + m
}

func minsToTimeStr(m int) string {
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

func sortSlots(slots []model.Slot) {
	for i := 1; i < len(slots); i++ {
		for j := i; j > 0 && slots[j-1].Time > slots[j].Time; j-- {
			slots[j-1], slots[j] = slots[j], slots[j-1]
		}
	}
}
