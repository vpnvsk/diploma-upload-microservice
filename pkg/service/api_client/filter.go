package api_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"reflect"
	"strings"
	"sync"
)

var returnFields = []string{
	"legal_status",
	"current_assignee",
	"inventor",
	"titles",
	"current_owner",
	"expiration_date",
	"assignee",
	"priority_claims",
	"app_pub_references",
}

type FilteredResponse struct {
	mu       sync.Mutex
	response *[]model.FilteredPatent
}

func (c *APIClient) FilterPatents(req model.Filters) (*model.FilteredPatentsResponse, error) {
	limit := *req.Limit
	offset := *req.Offset
	parsedResponse := make([]model.FilteredPatent, 0, limit)

	response := FilteredResponse{
		mu:       sync.Mutex{},
		response: &parsedResponse,
	}
	parsedFilters, err := c.parseFilters(req)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	for i := 0; i < limit; i = i + 5 {
		wg.Add(1)
		go func(i int) {
			local := make([]model.FilteredPatent, 0, chunkSize)
			defer wg.Done()
			c.getFilteredChunk(parsedFilters, offset*limit+i, req.PreFilter, &local)
			response.mu.Lock()
			*response.response = append(*response.response, local...)
			response.mu.Unlock()
		}(i)
	}

	responseWithStats := &model.FilteredPatentsResponse{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.getStatistics(parsedFilters, responseWithStats)
	}()
	wg.Wait()
	responseWithStats.Patents = response.response
	return responseWithStats, err
}

func (c *APIClient) getFilteredChunk(parsedFilters []model.SingleParsedFilter, offset int, preFilter *bool, response *[]model.FilteredPatent) {
	patents, err := c.repo.GetFilteredData(
		model.NewFilterRequestBody(parsedFilters, c.cfg.KTMineAPIKey, offset, 5, preFilter, returnFields),
	)
	if err != nil {
		fmt.Println(err)
	}
	if err = c.parseFilteredResponse(patents, response); err != nil {
		fmt.Println(err)
	}
}

func (c *APIClient) getStatistics(parsedFilters []model.SingleParsedFilter, parsedResponse *model.FilteredPatentsResponse) {
	response, err := c.repo.GetFilteredData(model.NewStatisticsRequestBody(parsedFilters, c.cfg.KTMineAPIKey))
	if err != nil {
		fmt.Println(err)
	}

	result, totalPatents, err := c.parseStatistics(response)
	if err != nil {
		fmt.Println(err)
	}

	parsedResponse.Statistics = result
	parsedResponse.TotalPatents = totalPatents
}

func (c *APIClient) parseFilteredResponse(patents *[]byte, parsedResponse *[]model.FilteredPatent) error {
	var data map[string]interface{}
	err := json.Unmarshal(*patents, &data)
	if err != nil {
		return err
	}
	response, ok := data["response"].(map[string]interface{})
	if !ok {
		return errors.New("can't parse response body")
	}
	items, ok := response["items"].([]interface{})
	for _, patent := range items {
		parsedPatent := c.parseFilteredPatent(patent)
		*parsedResponse = append(*parsedResponse, parsedPatent)
	}
	return nil
}

func (c *APIClient) parseTermsFilters(termFilter string) string {
	var builder strings.Builder

	fields := []string{
		"invention_title",
		"descriptions.plain_text",
		"claims.plain_text",
		"abstract_paragraphs.plain_text",
	}

	builder.WriteString("(")

	for i, field := range fields {
		if i > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString(fmt.Sprintf("%s:(%s)", field, termFilter))
	}

	builder.WriteString(")")

	return builder.String()
}

func (c *APIClient) parseFilters(filters model.Filters) ([]model.SingleParsedFilter, error) {
	var parsedFilters []model.SingleParsedFilter

	if filters.TermsFilters != nil {
		parsedFilters = append(parsedFilters, *model.NewSingleParsedFilter([]string{c.parseTermsFilters(*filters.TermsFilters)}, "fulltext", nil))
	}

	v := reflect.ValueOf(filters)
	t := reflect.TypeOf(filters)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if field.Name == "Limit" || field.Name == "Offset" || field.Name == "PreFilter" || field.Name == "TermsFilterSimplified" {
			continue
		}

		if fieldValue.IsNil() {
			continue
		}

		key := field.Name

		switch fieldValue.Interface().(type) {
		case *[]model.DateInFilter:
			dateFilters := fieldValue.Interface().(*[]model.DateInFilter)
			if len(*dateFilters) > 0 {
				dateFiltersParsed := *dateFilters
				minStr := dateFiltersParsed[0].Min.Format("2006-01-02")
				maxStr := dateFiltersParsed[0].Max.Format("2006-01-02")

				parsedFilters = append(parsedFilters, *model.NewSingleParsedFilter(
					[]string{minStr, maxStr}, key, nil))
			}

		case *[]model.SingleFilter:
			singleFilters := fieldValue.Interface().(*[]model.SingleFilter)
			if len(*singleFilters) > 0 {
				var criteria []string
				for _, filter := range *singleFilters {
					criteria = append(criteria, filter.Value)
				}
				parsedFilters = append(parsedFilters, *model.NewSingleParsedFilter(criteria, key, nil))
			}

		case *[]string:
			stringFilters := fieldValue.Interface().(*[]string)
			if len(*stringFilters) > 0 {
				operator := model.OrOperator
				parsedFilters = append(parsedFilters, *model.NewSingleParsedFilter(*stringFilters, key, &operator))
			}
		}
	}

	return parsedFilters, nil
}
