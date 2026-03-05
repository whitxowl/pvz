package domain

import "time"

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

type Type string

const (
	TypeElectronics Type = "элекктроника"
	TypeClothes     Type = "одежда"
	TypeShoes       Type = "обувь"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeElectronics, TypeClothes, TypeShoes:
		return true
	default:
		return false
	}
}

type Reception struct {
	ID       string
	Date     *time.Time
	PvzID    string
	Status   Status
	Products []Product
}

type Product struct {
	ID          string
	Date        *time.Time
	Type        Type
	ReceptionID string
}
