package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

const confirmationTmpl = `RE:FORM — Booking Confirmed

Hi {{.CustomerName}},

Your appointment is confirmed.

  Barber:  {{.BarberName}}
  Service: {{.ServiceName}}
  Date:    {{.Date}}
  Time:    {{.Time}}
  Ref:     {{.Reference}}

3 Upper Main Street, Larne, BT40 · Northern Ireland
`

type EmailNotifier struct {
	apiKey    string
	fromEmail string
	fromName  string
	tmpl      *template.Template
}

func NewEmailNotifier(apiKey, fromEmail, fromName string) *EmailNotifier {
	return &EmailNotifier{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		tmpl:      template.Must(template.New("confirm").Parse(confirmationTmpl)),
	}
}

func (e *EmailNotifier) BookingConfirmed(ctx context.Context, p BookingPayload) error {
	var buf bytes.Buffer
	if err := e.tmpl.Execute(&buf, p); err != nil {
		return err
	}
	return e.send(ctx, p.CustomerEmail, "RE:FORM — Booking Confirmed · "+p.Reference, buf.String())
}

func (e *EmailNotifier) AppointmentReminder(ctx context.Context, p BookingPayload) error {
	body := fmt.Sprintf("Reminder: your appointment with %s is tomorrow at %s. Ref: %s", p.BarberName, p.Time, p.Reference)
	return e.send(ctx, p.CustomerEmail, "RE:FORM — Appointment Tomorrow", body)
}

func (e *EmailNotifier) send(ctx context.Context, to, subject, text string) error {
	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", e.fromName, e.fromEmail),
		"to":      []string{to},
		"subject": subject,
		"text":    text,
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend: unexpected status %d", resp.StatusCode)
	}
	return nil
}
