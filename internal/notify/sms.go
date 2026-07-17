package notify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type SMSNotifier struct {
	accountSID string
	authToken  string
	fromNumber string
}

func NewSMSNotifier(accountSID, authToken, fromNumber string) *SMSNotifier {
	return &SMSNotifier{accountSID: accountSID, authToken: authToken, fromNumber: fromNumber}
}

func (s *SMSNotifier) BookingConfirmed(ctx context.Context, p BookingPayload) error {
	if p.CustomerPhone == "" {
		return nil
	}
	msg := fmt.Sprintf("RE:FORM: Confirmed - %s with %s on %s at %s. Ref: %s", p.ServiceName, p.BarberName, p.Date, p.Time, p.Reference)
	return s.send(ctx, p.CustomerPhone, msg)
}

func (s *SMSNotifier) AppointmentReminder(ctx context.Context, p BookingPayload) error {
	if p.CustomerPhone == "" {
		return nil
	}
	msg := fmt.Sprintf("RE:FORM: Reminder - %s with %s tomorrow at %s. Ref: %s", p.ServiceName, p.BarberName, p.Time, p.Reference)
	return s.send(ctx, p.CustomerPhone, msg)
}

func (s *SMSNotifier) send(ctx context.Context, to, body string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.accountSID)
	v := url.Values{}
	v.Set("To", to)
	v.Set("From", s.fromNumber)
	v.Set("Body", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.accountSID, s.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("twilio: unexpected status %d", resp.StatusCode)
	}
	return nil
}
