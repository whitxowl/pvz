package domain

import "time"

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

type Product struct {
	ID          string
	Date        *time.Time
	Type        Type
	ReceptionID string
}
