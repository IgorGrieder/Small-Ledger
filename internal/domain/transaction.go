package domain

import "github.com/google/uuid"

type Transaction struct {
	From          string
	To            string
	Currency      string
	Value         string
	CorrelationId uuid.UUID
}
