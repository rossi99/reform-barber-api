// Package model defines the JSON shapes returned by the API.
// These are deliberately separate from the sqlc-generated DB types.
package model

import (
	"time"

	"github.com/google/uuid"
)

type Barber struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Bio      string    `json:"bio"`
	Num      string    `json:"num"`
	PhotoURL string    `json:"photo_url,omitempty"`
}

type Service struct {
	ID           uuid.UUID `json:"id"`
	Num          string    `json:"num"`
	Name         string    `json:"name"`
	NameHTML     string    `json:"name_html"`
	Description  string    `json:"description"`
	DurationMins int32     `json:"duration"`
	PricePence   int32     `json:"price"`
}

type Product struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	PricePence int32     `json:"price"`
	ImageURL   string    `json:"image_url,omitempty"`
}

type Slot struct {
	Time   string `json:"time"`
	Booked bool   `json:"booked"`
}

type AvailabilityResponse struct {
	Date   string `json:"date"`
	Groups struct {
		Morning   []Slot `json:"morning"`
		Afternoon []Slot `json:"afternoon"`
		Evening   []Slot `json:"evening"`
	} `json:"groups"`
}

type BookingRequest struct {
	BarberID  string            `json:"barberId"`  // UUID or "any"
	ServiceID string            `json:"serviceId"` // UUID
	Date      string            `json:"date"`      // YYYY-MM-DD
	Time      string            `json:"time"`      // HH:MM
	Products  map[string]int    `json:"products"`  // productID → qty
}

type BookingResponse struct {
	ID        uuid.UUID `json:"id"`
	Reference string    `json:"reference"`
	Status    string    `json:"status"`
	BarberID  uuid.UUID `json:"barberId"`
	ServiceID uuid.UUID `json:"serviceId"`
	Date      string    `json:"date"`
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"createdAt"`
}

type MediaItem struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	PublicURL string    `json:"public_url"`
	AltText   string    `json:"alt_text,omitempty"`
	SortOrder int32     `json:"sort_order"`
}

type UserClaims struct {
	UserID   uuid.UUID
	Role     string
	BarberID *uuid.UUID
}
