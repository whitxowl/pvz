package domain

import "time"

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "close"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

type Reception struct {
	ID     string
	Date   *time.Time
	PvzID  string
	Status Status
}
