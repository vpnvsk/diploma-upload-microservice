package model

import (
	"strings"
)

type UploadFilterOperator string

const (
	AndOperator UploadFilterOperator = "and"
	OrOperator  UploadFilterOperator = "or"
)

type SingleFilter struct {
	Criteria       []string             `json:"criteria"`
	SearchField    string               `json:"searchField"`
	FilterOperator UploadFilterOperator `json:"filterOperator"`
}

func NewSingleFilter(criteria []string, searchField string, operator UploadFilterOperator) *SingleFilter {
	if operator == "" {
		operator = AndOperator
	}

	filter := &SingleFilter{
		Criteria:       criteria,
		SearchField:    searchField,
		FilterOperator: operator,
	}
	filter.SearchFieldValidator()
	return filter

}

func (s *SingleFilter) SearchFieldValidator() {
	parsedName := strings.ReplaceAll(s.SearchField, "_", "")
	s.SearchField = "patent." + strings.ToLower(parsedName)
}

type FiltersRequestBody struct {
	Filters       []SingleFilter `json:"filters"`
	ReturnFields  []string       `json:"returnFields"`
	Key           string         `json:"key"`
	SortField     string         `json:"sortField"`
	SortDirection string         `json:"sortDirection"`
	Start         int            `json:"start"`
	Count         int            `json:"count"`
}

func NewFilterRequestBody(filters []SingleFilter, key string, start, count int) FiltersRequestBody {
	return FiltersRequestBody{
		Filters:       filters,
		ReturnFields:  []string{"all"},
		Key:           key,
		SortField:     "docdb_document_number",
		SortDirection: "asc",
		Start:         start,
		Count:         count,
	}
}
