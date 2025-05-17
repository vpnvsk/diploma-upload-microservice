package model

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type UploadInput struct {
	PublicationNumbers []string `json:"publication_numbers" validate:"required"`
	CollectionId       uuid.UUID
	UserId             uuid.UUID
}

type CustomDate struct {
	time.Time
}

func (c *CustomDate) UnmarshalJSON(b []byte) error {
	layout := `"2006-01-02"`
	parsedTime, err := time.Parse(layout, string(b))
	if err != nil {
		return fmt.Errorf("invalid date format: %v", err)
	}
	c.Time = parsedTime
	return nil
}

type DateInFilter struct {
	Min CustomDate `json:"min,omitempty"`
	Max CustomDate `json:"max,omitempty"`
}

type SingleFilter struct {
	Value    string                `json:"value"`
	Operator *UploadFilterOperator `json:"operator"`
}

type Filters struct {
	DocumentNumber  *[]string       `json:"document_number,omitempty"`
	ApplicationDate *[]DateInFilter `json:"application_date,omitempty"`
	PublicationDate *[]DateInFilter `json:"publication_date,omitempty"`
	CurrentAssignee *[]SingleFilter `json:"current_assignee,omitempty"`
	Inventor        *[]SingleFilter `json:"inventor,omitempty"`
	CurrentOwner    *[]SingleFilter `json:"current_owner,omitempty"`
	DocumentCountry *[]SingleFilter `json:"document_country,omitempty"`
	TopLevelCPC     *[]SingleFilter `json:"top_level_cpc,omitempty"`
	CPCCode         *[]SingleFilter `json:"cpc_code,omitempty"`
	LegalStatus     *[]SingleFilter `json:"legal_status,omitempty"`
	TermsFilters    *string         `json:"terms_filter_simplified,omitempty"`
	PreFilter       *bool           `json:"pre_filter"`
	Limit           *int            `json:"limit"`
	Offset          *int            `json:"offset"`
}

func (f *Filters) Sanitize() {
	if f.Offset == nil {
		offset := 0
		f.Offset = &offset
	}
	if f.Limit == nil {
		limit := 10
		f.Limit = &limit
	}
}
