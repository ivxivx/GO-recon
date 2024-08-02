package domain

import "time"

type Transaction interface {
	GetMatchingKey() string

	GetID() string
	GetExternalID() *string
	GetType() string
	GetTimestamp() time.Time
}
