package model

import (
	"strings"
)

type UploadFilterOperator string

var advancedAggs = []map[string]string{
	{"field": "current_assignee.party_name.raw", "name": "Current Assignee", "type": "terms"},
	{"field": "document_country", "name": "Document Country", "type": "terms"},
	{"field": "legal_status", "name": "Legal Status", "type": "terms"},
	{"field": "classifications_cpc.section_top_class_sub_class", "name": "Top Level CPCs", "type": "terms"},
}

const (
	AndOperator UploadFilterOperator = "and"
	OrOperator  UploadFilterOperator = "or"
	NotOperator UploadFilterOperator = "not"
)

type SingleParsedFilter struct {
	Criteria       []string              `json:"criteria"`
	SearchField    string                `json:"searchField"`
	FilterOperator *UploadFilterOperator `json:"filterOperator"`
}

func NewSingleParsedFilter(criteria []string, searchField string, operator *UploadFilterOperator) *SingleParsedFilter {
	if operator == nil {
		defaultOperator := AndOperator
		operator = &defaultOperator
	}

	filter := &SingleParsedFilter{
		Criteria:       criteria,
		SearchField:    searchField,
		FilterOperator: operator,
	}
	filter.SearchFieldValidator()
	return filter
}

func (s *SingleParsedFilter) SearchFieldValidator() {
	parsedName := strings.ReplaceAll(s.SearchField, "_", "")
	s.SearchField = "patent." + strings.ToLower(parsedName)
}

type FilterInterface interface {
	GetFilters() []SingleParsedFilter
}

type FiltersRequestBody struct {
	Filters      []SingleParsedFilter `json:"filters"`
	ReturnFields []string             `json:"returnFields"`
	Key          string               `json:"key"`
	Start        int                  `json:"start"`
	Count        int                  `json:"count"`
	PreFilter    *bool                `json:"preFilter,omitempty"`
}

func (f FiltersRequestBody) GetFilters() []SingleParsedFilter { return f.Filters }

func NewFilterRequestBody(
	filters []SingleParsedFilter, key string, start, count int, preFilter *bool, returnFields []string) FiltersRequestBody {
	var preFilterParsed bool
	if preFilter != nil {
		preFilterParsed = *preFilter
	} else {
		preFilterParsed = false
	}
	return FiltersRequestBody{
		Filters:      filters,
		ReturnFields: returnFields,
		Key:          key,
		Start:        start,
		Count:        count,
		PreFilter:    &preFilterParsed,
	}
}

type StatisticsRequestBody struct {
	Filters      []SingleParsedFilter `json:"filters"`
	Start        int                  `json:"start"`
	Count        int                  `json:"count"`
	Key          string               `json:"key"`
	AdvancedAggs []map[string]string  `json:"advancedAggs"`
}

func (s StatisticsRequestBody) GetFilters() []SingleParsedFilter { return s.Filters }

func NewStatisticsRequestBody(filters []SingleParsedFilter, key string) StatisticsRequestBody {
	return StatisticsRequestBody{
		Filters:      filters,
		Start:        0,
		Count:        0,
		Key:          key,
		AdvancedAggs: advancedAggs,
	}
}
