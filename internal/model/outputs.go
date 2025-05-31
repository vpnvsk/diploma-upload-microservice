package model

import "github.com/google/uuid"

type AnalyzePatentsOutput struct {
	TransactionId uuid.UUID  `json:"transaction_id"`
	BundleId      uuid.UUID  `json:"bundle_id"`
	UserId        *uuid.UUID `json:"user_id,omitempty"`
}
