package domain

import "time"

type City string

const (
	Moscow          City = "Москва"
	SaintPetersburg City = "Санкт-Петербург"
	Kazan           City = "Казань"
)

func (c City) IsValid() bool {
	switch c {
	case Moscow, SaintPetersburg, Kazan:
		return true
	default:
		return false
	}
}

type PVZ struct {
	ID               string
	RegistrationDate *time.Time
	City             City
}
