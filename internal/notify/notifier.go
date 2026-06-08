package notify

import "context"

type BookingPayload struct {
	Reference    string
	CustomerName string
	CustomerEmail string
	CustomerPhone string
	BarberName   string
	ServiceName  string
	Date         string // "Mon, 19 May"
	Time         string // "11:00"
}

type Notifier interface {
	BookingConfirmed(ctx context.Context, p BookingPayload) error
	AppointmentReminder(ctx context.Context, p BookingPayload) error
}

// Noop discards all notifications. Used in tests and when providers are not configured.
type Noop struct{}

func (Noop) BookingConfirmed(_ context.Context, _ BookingPayload) error   { return nil }
func (Noop) AppointmentReminder(_ context.Context, _ BookingPayload) error { return nil }
