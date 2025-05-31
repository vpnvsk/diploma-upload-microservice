package api_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"golang.org/x/sync/errgroup"
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

var fullPatentReturnFields = []string{
	"descriptions",
	"abstract",
	"images",
}

type FilteredResponse struct {
	mu       sync.Mutex
	response *[]model.FilteredPatent
}

func (c *APIClient) FilterPatents(ctx context.Context, req model.Filters) (*model.FilteredPatentsResponse, error) {
	limit := *req.Limit
	offset := *req.Offset
	g, ctx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	patents := make([]model.FilteredPatent, 0, limit)
	parsedFilters, err := c.ParseFilters(req)
	if err != nil {
		return nil, err
	}

	for i := 0; i < limit; i = i + 5 {
		off := i
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			local := make([]model.FilteredPatent, 0, chunkSize)

			err := c.getFilteredChunk(ctx, parsedFilters, offset*limit+off, req.PreFilter, &local)
			if err != nil {
				return fmt.Errorf("chunk @%d: %w", off, err)
			}

			mu.Lock()
			patents = append(patents, local...)
			mu.Unlock()
			return nil
		})
	}

	responseWithStats := &model.FilteredPatentsResponse{}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		statistics, totalPatent, err := c.GetStatistics(ctx, parsedFilters)
		if err != nil {
			return fmt.Errorf("error getting statistics: %w", err)
		}
		responseWithStats.TotalPatents = totalPatent
		responseWithStats.Statistics = statistics
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	responseWithStats.Patents = &patents
	return responseWithStats, err
}

func (c *APIClient) getFilteredChunk(
	ctx context.Context,
	parsedFilters []model.SingleParsedFilter,
	offset int,
	preFilter *bool,
	response *[]model.FilteredPatent,
) error {
	patents, err := c.repo.GetFilteredData(
		ctx, model.NewFilterRequestBody(parsedFilters, c.cfg.KTMineAPIKey, offset, 5, preFilter, returnFields),
	)
	if err != nil {
		return err
	}
	if err = c.parseFilteredResponse(patents, response); err != nil {
		return err
	}
	return nil
}

func (c *APIClient) GetFilteredChunkFullPatentRaw(
	ctx context.Context,
	parsedFilters []model.SingleParsedFilter,
	offset int,
	limit int,
) (*[]byte, error) {
	returnFieldsFull := append(returnFields, fullPatentReturnFields...)
	patents, err := c.repo.GetFilteredData(
		ctx, model.NewFilterRequestBody(parsedFilters, c.cfg.KTMineAPIKey, offset, limit, nil, returnFieldsFull),
	)
	return patents, err
}

func (c *APIClient) GetStatistics(
	ctx context.Context,
	parsedFilters []model.SingleParsedFilter,
) (*map[string]interface{}, int, error) {
	response, err := c.repo.GetFilteredData(ctx, model.NewStatisticsRequestBody(parsedFilters, c.cfg.KTMineAPIKey))
	if err != nil {
		return nil, 0, err
	}

	result, totalPatents, err := c.parseStatistics(response)
	if err != nil {
		return nil, 0, err
	}
	return result, totalPatents, nil
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

func (c *APIClient) ParseFilters(filters model.Filters) ([]model.SingleParsedFilter, error) {
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
