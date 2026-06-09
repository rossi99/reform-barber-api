package main

import (
	"strconv"

	"github.com/reform-barber/api/internal/notify"
)

func buildNotifier() notify.Notifier {
	var notifiers []notify.Notifier

	emailsEnabled, err := strconv.ParseBool(mustEnv("EMAIL_NOTIFICATIONS_ENABLED"))
	if err != nil {
		logger.Warn().Msg("failed to parse value for email notifications enabled, setting to false")
		emailsEnabled = false
	}

	smsEnabled, err := strconv.ParseBool(mustEnv("SMS_NOTIFICATIONS_ENABLED"))
	if err != nil {
		logger.Warn().Msg("failed to parse value for text notifications enabled, setting to false")
		smsEnabled = false
	}

	emailNotifier := initEmailNotification(emailsEnabled)
	smsNotifier := initTxtNotifications(smsEnabled)

	if emailNotifier != nil {
		logger.Info().Msg("email notifications enabled")
		notifiers = append(notifiers, emailNotifier)
	}
	if smsNotifier != nil {
		logger.Info().Msg("sms notifications enabled")
		notifiers = append(notifiers, smsNotifier)
	}

	if len(notifiers) == 0 {
		logger.Warn().Msg("no notification providers configured — using noop")
		return notify.Noop{}
	}
	return notify.NewMulti(notifiers...)
}

func initEmailNotification(emailsEnabled bool) *notify.EmailNotifier {
	if emailsEnabled {
		return notify.NewEmailNotifier(mustEnv("RESEND_API_KEY"), mustEnv("EMAIL_FROM"), "RE:FORM")
	}
	return nil
}

func initTxtNotifications(smsEnabled bool) *notify.SMSNotifier {
	if smsEnabled {
		return notify.NewSMSNotifier(mustEnv("TWILIO_ACCOUNT_SID"), mustEnv("TWILIO_AUTH_TOKEN"), mustEnv("TWILIO_FROM_NUMBER"))
	}
	return nil
}
