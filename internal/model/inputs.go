package model

import "github.com/google/uuid"

type UploadInput struct {
	PublicationNumbers []string `json:"publication_numbers" validate:"required"`
	CollectionId       uuid.UUID
	UserId             uuid.UUID
}
