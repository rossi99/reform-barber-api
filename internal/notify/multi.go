package notify

import (
	"context"
	"errors"
)

// Multi fans out to multiple Notifier implementations, collecting all errors.
type Multi struct {
	notifiers []Notifier
}

func NewMulti(nn ...Notifier) *Multi { return &Multi{notifiers: nn} }

func (m *Multi) BookingConfirmed(ctx context.Context, p BookingPayload) error {
	return m.fan(func(n Notifier) error { return n.BookingConfirmed(ctx, p) })
}

func (m *Multi) AppointmentReminder(ctx context.Context, p BookingPayload) error {
	return m.fan(func(n Notifier) error { return n.AppointmentReminder(ctx, p) })
}

func (m *Multi) fan(fn func(Notifier) error) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := fn(n); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
