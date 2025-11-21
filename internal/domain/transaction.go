package domain

import "github.com/google/uuid"

type Transaction struct {
	From          uuid.UUID
	To            uuid.UUID
	Currency      string
	Value         string
	CorrelationId uuid.UUID
}
