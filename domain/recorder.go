package domain

import "time"

type Activity struct {
	UserID    string
	Action    string
	Timestamp time.Time
	Metadata  map[string]interface{}
}
