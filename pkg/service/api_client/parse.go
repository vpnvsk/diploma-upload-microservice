package api_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"time"
)

func (c *APIClient) parseInventors(payload []interface{}) []string {
	parsedInventors := make(map[string]struct{})
	for _, inventor := range payload {
		parsedInventor, _ := inventor.(map[string]interface{})
		var partyName string
		if partyName, _ = parsedInventor["partyNameClean"].(string); partyName == "" {
			partyName, _ = parsedInventor["partyName"].(string)
		}
		if partyName != "" {
			parsedInventors[partyName] = struct{}{}
		}
	}
	uniqueInventors := make([]string, 0, len(parsedInventors))
	for inventor := range parsedInventors {
		uniqueInventors = append(uniqueInventors, inventor)
	}
	return uniqueInventors
}

func (c *APIClient) parseAssignees(payload []interface{}) []string {
	parsedAssignees := make(map[string]struct{})
	for _, assignee := range payload {
		parsedAssignee, _ := assignee.(map[string]interface{})
		var partyName string
		if partyName, _ = parsedAssignee["partyNameClean"].(string); partyName == "" {
			partyName, _ = parsedAssignee["partyName"].(string)
		}

		if partyName != "" {
			parsedAssignees[partyName] = struct{}{}
		}
	}
	parsedUniqueAssignees := make([]string, 0, len(parsedAssignees))
	for assignee := range parsedAssignees {
		parsedUniqueAssignees = append(parsedUniqueAssignees, assignee)
	}
	return parsedUniqueAssignees
}

func (c *APIClient) parseSimpleFamilyJurisdiction(minPriorityDate string, payload []interface{}) []string {
	parsedFamilyJurisdiction := make(map[string]struct{})
	for _, priorityClaim := range payload {
		parsedPriorityClaim, _ := priorityClaim.(map[string]interface{})
		documentDate, _ := parsedPriorityClaim["documentDate"].(string)
		if documentDate == minPriorityDate {
			country, _ := parsedPriorityClaim["country"].(string)
			if country != "" {
				parsedFamilyJurisdiction[country] = struct{}{}
			}
		}
	}
	uniqueJurisdictions := make([]string, 0, len(parsedFamilyJurisdiction))
	for jurisdiction := range parsedFamilyJurisdiction {
		uniqueJurisdictions = append(uniqueJurisdictions, jurisdiction)
	}
	return uniqueJurisdictions
}

func (c *APIClient) parseFilteredPatent(patent interface{}) model.FilteredPatent {
	parsedPatent, _ := patent.(map[string]interface{})
	publicationNumber, _ := parsedPatent["documentNumber"].(string)
	simpleLegalStatus, _ := parsedPatent["legalStatus"].(string)

	title, ok := parsedPatent["inventionTitle"].(string)
	if !ok || title == "" {
		inventionTitles, ok := parsedPatent["inventionTitles"].([]interface{})
		if ok {
			for _, it := range inventionTitles {
				inventionTitle, ok := it.(map[string]interface{})
				if ok {
					if lang, ok := inventionTitle["lang"].(string); ok && lang == "eng" {
						title, _ = inventionTitle["title"].(string)
						break
					}
				}
			}
		}
	}
	var applicationDate string
	if applicationReferences, ok := parsedPatent["applicationReferences"].([]interface{}); ok {
		for _, applicationReference := range applicationReferences {
			if applicationReferenceParsed, ok := applicationReference.(map[string]interface{}); ok {
				if applicationDate == "" {
					applicationDate, ok = applicationReferenceParsed["documentDate"].(string)
					if !ok {
						applicationDate = ""
					}
				}
			}
		}
	}

	const customDateLayout = "2006-01-02T15:04:05"
	applicationDateParsed, err := time.Parse(customDateLayout, applicationDate)
	if err != nil {
		applicationDateParsed = time.Time{}
	}
	earliestPriorityDate, _ := parsedPatent["minPriorityDate"].(string)
	estimatedExpiryDate, _ := parsedPatent["projectedExpirationDate"].(string)
	earliestPriorityDateParsed, err := time.Parse(customDateLayout, earliestPriorityDate)
	if err != nil {
		earliestPriorityDateParsed = time.Time{}
	}
	estimatedExpiryDateParsed, err := time.Parse(customDateLayout, estimatedExpiryDate)
	if err != nil {
		estimatedExpiryDateParsed = time.Time{}
	}
	minPriorityDate, _ := parsedPatent["minPriorityDate"].(string)

	priorityClaims, _ := parsedPatent["priorityClaims"].([]interface{})
	uniqueJurisdictions := c.parseSimpleFamilyJurisdiction(minPriorityDate, priorityClaims)

	var currentAssignee []interface{}
	currentAssignee, _ = parsedPatent["currentOwners"].([]interface{})
	if len(currentAssignee) == 0 {
		currentAssignee, _ = parsedPatent["currentAssignees"].([]interface{})
		if len(currentAssignee) == 0 {
			currentAssignee, _ = parsedPatent["assignees"].([]interface{})
		}
	}
	parsedUniqueAssignees := c.parseAssignees(currentAssignee)

	inventors, _ := parsedPatent["inventors"].([]interface{})
	uniqueInventors := c.parseInventors(inventors)

	return model.FilteredPatent{
		Title:                    title,
		PublicationNumber:        publicationNumber,
		EarliestPriorityDate:     &earliestPriorityDateParsed,
		EstimatedExpiryDate:      &estimatedExpiryDateParsed,
		InventorsNames:           uniqueInventors,
		Assignee:                 parsedUniqueAssignees,
		SimpleFamilyJurisdiction: uniqueJurisdictions,
		ApplicationDate:          &applicationDateParsed,
		SimpleLegalStatus:        &simpleLegalStatus,
	}
}
func (c *APIClient) parseStatistics(payload *[]byte) (*map[string]interface{}, int, error) {
	var data map[string]interface{}
	err := json.Unmarshal(*payload, &data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	response, ok := data["response"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("missing or invalid 'response' field")
	}

	stats, ok := data["aggregations"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("missing or invalid 'aggregations' field")
	}

	// Ensure correct type conversion for totalPatents
	totalFound, ok := response["totalFound"]
	if !ok {
		return nil, 0, errors.New("missing 'totalFound' in response")
	}

	totalPatents, ok := totalFound.(float64) // JSON numbers are float64 by default
	if !ok {
		return nil, 0, fmt.Errorf("invalid 'totalFound' type: expected float64, got %T", totalFound)
	}

	return &stats, int(totalPatents), nil
}
