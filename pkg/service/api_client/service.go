package api_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/internal/utils"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const chunkSize = 20

type APIClient struct {
	cfg  *config.Config
	log  *slog.Logger
	repo repository.KTMineRepositoryInterface
}

func NewAPIClient(log *slog.Logger, repo repository.KTMineRepositoryInterface, cfg *config.Config) *APIClient {
	return &APIClient{
		cfg:  cfg,
		log:  log,
		repo: repo,
	}
}

func (c *APIClient) GetData(input model.UploadInput) error {
	var wg sync.WaitGroup
	packagesChan := make(chan *model.ParsedPatentsData, 1000)
	for i := 0; i < len(input.PublicationNumbers); i += chunkSize {
		end := i + chunkSize
		if end > len(input.PublicationNumbers) {
			end = len(input.PublicationNumbers)
		}
		wg.Add(1)
		requestBody := model.NewFilterRequestBody([]model.SingleFilter{*model.NewSingleFilter(
			input.PublicationNumbers[i:end], "documentnumber", "and")},
			c.cfg.KTMineAPIKey, 0, chunkSize)
		go func() error {
			defer wg.Done()
			if err := c.getAndParseData(requestBody, packagesChan); err != nil {
				fmt.Println(err)
			}
			return nil
		}()
	}
	wg.Wait()
	close(packagesChan)

	return nil
}

func (c *APIClient) getAndParseData(requestBody model.FiltersRequestBody, ch chan *model.ParsedPatentsData) error {
	body, err := c.sendRequest(requestBody)
	if err != nil {
		return err
	}
	err = c.parseResponse(body, ch)

	return err
}

func (c *APIClient) sendRequest(requestBody model.FiltersRequestBody) (*[]byte, error) {
	var body *[]byte
	var err error
	for i := 0; i < 3; i++ {
		body, err = c.repo.GetFilteredData(requestBody)
		if err != nil {
			continue
		} else {
			return body, err
		}
	}
	return body, err
}

func (c *APIClient) parseResponse(body *[]byte, ch chan *model.ParsedPatentsData) error {
	var data map[string]interface{}
	err := json.Unmarshal(*body, &data)
	if err != nil {
		return err
	}
	response, ok := data["response"].(map[string]interface{})
	if !ok {
		return errors.New("can't parse response body")
	}
	items, ok := response["items"].([]interface{})
	if !ok {
		return errors.New("can't parse response body")
	}
	fmt.Println(len(items))

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return errors.New("can't parse response body")
		}
		parsedPatent := c.parsePatent(itemMap)
		parsedInventors, parsedInventorPatentLink := c.parseInventors(itemMap, parsedPatent.Id)
		parsedAssignees, parsedAssigneePatentLink := c.parseAssignees(itemMap, parsedPatent.Id)
		parsedJurisdictions, parsedJurisdictionsPatentLink := c.parseJurisdictions(itemMap, parsedPatent.Id,
			parsedPatent.EarliestPriorityDate)
		parsedClaims := c.parseClaim(itemMap, parsedPatent.Id)

		ch <- &model.ParsedPatentsData{
			Patent:                  parsedPatent,
			Inventors:               parsedInventors,
			PatentInventorsLink:     parsedInventorPatentLink,
			Assignees:               parsedAssignees,
			PatentAssigneesLink:     parsedAssigneePatentLink,
			Jurisdictions:           parsedJurisdictions,
			PatentJurisdictionsLink: parsedJurisdictionsPatentLink,
			Claims:                  parsedClaims,
		}
	}

	return nil
}

func (c *APIClient) parseInventors(data map[string]interface{}, patentId uuid.UUID) (
	*[]model.Inventor, *[]model.PatentInventorLink) {
	inventors := make([]model.Inventor, 0)
	inventorPatentLink := make([]model.PatentInventorLink, 0)
	inventorsList, ok := data["inventors"].([]interface{})
	if ok {
		for _, name := range inventorsList {
			parsedInventorName, ok := name.(map[string]interface{})
			if ok {
				if parsedName, ok := parsedInventorName["partyNameClean"].(string); ok {
					inventors = append(inventors, model.Inventor{FullName: parsedName})
					inventorPatentLink = append(inventorPatentLink, model.PatentInventorLink{
						PatentID: patentId, InventorName: parsedName})
				} else if parsedName, ok := parsedInventorName["partyName"].(string); ok {
					inventors = append(inventors, model.Inventor{FullName: parsedName})
					inventorPatentLink = append(inventorPatentLink, model.PatentInventorLink{
						PatentID: patentId, InventorName: parsedName})
				}
			}
		}
	}
	return &inventors, &inventorPatentLink
}

func (c *APIClient) parsePatent(data map[string]interface{}) *model.ParsedPatent {
	publicationNumber, _ := data["documentNumber"].(string)
	simpleLegalStatus, _ := data["legalStatus"].(string)

	title, ok := data["inventionTitle"].(string)
	if !ok || title == "" {
		inventionTitles, ok := data["inventionTitles"].([]interface{})
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

	var cpcList []string
	if cpcClassifications, ok := data["cpcClassifications"].([]interface{}); ok {
		for _, entry := range cpcClassifications {
			if classificationMap, ok := entry.(map[string]interface{}); ok {
				if symbol, ok := classificationMap["symbol"].(string); ok {
					cpcList = append(cpcList, symbol)
				}
			}
		}
	}
	cpcResult := strings.Join(cpcList, " | ")

	var inpadocFamilyMembers, inpadocFamilyJurisdictions []string
	if inpadocFamilyList, ok := data["inpadocFamilyMembers"].([]interface{}); ok {
		for _, entry := range inpadocFamilyList {
			if inpadocFamily, ok := entry.(map[string]interface{}); ok {
				country, _ := inpadocFamily["country"].(string)
				documentNumber, _ := inpadocFamily["documentNumber"].(string)
				kind, _ := inpadocFamily["kind"].(string)
				inpadocFamilyMember := fmt.Sprintf("%s%s%s", country, documentNumber, kind)
				inpadocFamilyMembers = append(inpadocFamilyMembers, inpadocFamilyMember)
				if country != "" {
					inpadocFamilyJurisdictions = append(inpadocFamilyJurisdictions, country)
				}
			}
		}
	}
	inpadocFamilyMembersResult := strings.Join(inpadocFamilyMembers, " | ")
	inpadocFamilyJurisdictionsResult := strings.Join(inpadocFamilyJurisdictions, " | ")

	var abstractResult string
	if abstractParagraph, ok := data["abstractParagraphs"].([]interface{}); ok {
		for _, abstract := range abstractParagraph {
			if abstractParsed, ok := abstract.(map[string]interface{}); ok {
				if abstractText, ok := abstractParsed["plainText"].(string); ok {
					abstractResult = fmt.Sprintf("%s%s\n", abstractResult, abstractText)
				}
			}
		}
	}
	abstractResult = utils.RemoveHTMLTags(abstractResult)

	briefDescriptionFlag := false
	var descriptionResult, briefDescriptionOfDrawingsResult string
	if descriptionData, ok := data["descriptions"].([]interface{}); ok {
		for _, descriptionText := range descriptionData {
			if description, ok := descriptionText.(map[string]interface{}); ok {
				var category string
				if category, ok = description["category"].(string); ok {
					if category == "brief-description-of-drawings" {
						briefDescriptionFlag = !briefDescriptionFlag
					}
				}
				if parsedDescription, ok := description["plainText"].(string); ok {
					descriptionResult = fmt.Sprintf("%s%s\n", descriptionResult, parsedDescription)
					if briefDescriptionFlag || category == "description-of-drawings" {
						briefDescriptionOfDrawingsResult = fmt.Sprintf("%s%s\n", briefDescriptionOfDrawingsResult,
							parsedDescription)
					}
				}
			}
		}
	}
	descriptionResult = utils.RemoveHTMLTags(descriptionResult)
	briefDescriptionOfDrawingsResult = utils.RemoveHTMLTags(briefDescriptionOfDrawingsResult)

	authority, _ := data["publicationReference"].(map[string]interface{})["country"].(string)

	var applicationNumber, applicationDate string
	if applicationReferences, ok := data["applicationReferences"].([]interface{}); ok {
		for _, applicationReference := range applicationReferences {
			if applicationReferenceParsed, ok := applicationReference.(map[string]interface{}); ok {
				if dataFormat, ok := applicationReferenceParsed["dataFormat"].(string); ok && dataFormat == "original" {
					if applicationNumber == "" {
						applicationNumber, ok = applicationReferenceParsed["documentNumber"].(string)
						if !ok {
							applicationNumber = ""
						}
					}
				}
				if applicationDate == "" {
					applicationDate, ok = applicationReferenceParsed["documentDate"].(string)
					if !ok {
						applicationDate = ""
					}
				}
			}
		}
	}
	applicationDateParsed, err := time.Parse("2006-01-02", applicationDate)
	if err != nil {
		applicationDateParsed = time.Time{}
	}

	var issueDate string
	if pubReferences, ok := data["publicationReferences"].([]interface{}); ok {
		for _, pubRef := range pubReferences {
			if pubRefMap, ok := pubRef.(map[string]interface{}); ok {
				if date, ok := pubRefMap["documentDate"].(string); ok && date != "" {
					issueDate = date
					break
				}
			}
		}
	}
	issueDateParsed, err := time.Parse("2006-01-02", issueDate)
	if err != nil {
		issueDateParsed = time.Time{}
	}

	earliestPriorityDate, _ := data["minPriorityDate"].(string)
	estimatedExpiryDate, _ := data["projectedExpirationDate"].(string)
	earliestPriorityDateParsed, err := time.Parse("2006-01-02", earliestPriorityDate)
	if err != nil {
		earliestPriorityDateParsed = time.Time{}
	}
	estimatedExpiryDateParsed, err := time.Parse("2006-01-02", estimatedExpiryDate)
	if err != nil {
		estimatedExpiryDateParsed = time.Time{}
	}
	var countOfCitedByPatents int
	if backwardCitations, ok := data["backwardCitations"].([]interface{}); ok && backwardCitations != nil {
		countOfCitedByPatents = len(backwardCitations)
	}

	return &model.ParsedPatent{
		Id:                             uuid.New(),
		Title:                          title,
		Abstract:                       abstractResult,
		CPC:                            cpcResult,
		EarliestPriorityDate:           earliestPriorityDateParsed,
		EstimatedExpiryDate:            estimatedExpiryDateParsed,
		PublicationNumber:              publicationNumber,
		CountOfCitedByPatents:          countOfCitedByPatents,
		Description:                    descriptionResult,
		BriefDescriptionOfDrawings:     briefDescriptionOfDrawingsResult,
		SimpleLegalStatus:              simpleLegalStatus,
		InpadocFamily:                  inpadocFamilyMembersResult,
		InpadocFamilyApplicationCount:  len(inpadocFamilyMembers),
		InpadocFamilyJurisdiction:      inpadocFamilyJurisdictionsResult,
		InpadocFamilyJurisdictionCount: len(inpadocFamilyJurisdictions),
		Authority:                      authority,
		ApplicationDate:                applicationDateParsed,
		ApplicationNumber:              applicationNumber,
		IssueDate:                      issueDateParsed,
		PublicationDate:                issueDateParsed,
		FileURL: fmt.Sprintf("https://api.ktmine.com/api/v2/patents/pdf/%s?key=%s",
			publicationNumber, c.cfg.KTMineAPIKey),
	}
}

func (c *APIClient) parseAssignees(data map[string]interface{}, patentId uuid.UUID) (
	*[]model.StandardizedCurrentAssignee, *[]model.PatentStandardizedCurrentAssigneeLink) {
	assignees := make([]model.StandardizedCurrentAssignee, 0)
	assigneePatentLink := make([]model.PatentStandardizedCurrentAssigneeLink, 0)
	assigneesList, ok := data["currentAssignees"].([]interface{})
	if ok {
		for _, name := range assigneesList {
			parsedInventorName, ok := name.(map[string]interface{})
			if ok {
				if parsedName, ok := parsedInventorName["partyNameClean"].(string); ok {
					assignees = append(assignees, model.StandardizedCurrentAssignee{Name: parsedName})
					assigneePatentLink = append(assigneePatentLink, model.PatentStandardizedCurrentAssigneeLink{
						PatentID: patentId, AssigneeName: parsedName})
				} else if parsedName, ok := parsedInventorName["partyName"].(string); ok {
					assignees = append(assignees, model.StandardizedCurrentAssignee{Name: parsedName})
					assigneePatentLink = append(assigneePatentLink, model.PatentStandardizedCurrentAssigneeLink{
						PatentID: patentId, AssigneeName: parsedName})
				}
			}
		}
	}
	return &assignees, &assigneePatentLink
}

func (c *APIClient) parseJurisdictions(data map[string]interface{}, patentId uuid.UUID, minPriorityDate time.Time) (
	*[]model.SimpleFamilyJurisdiction, *[]model.PatentSimpleFamilyJurisdictionLink) {

	jurisdictions := make([]model.SimpleFamilyJurisdiction, 0)
	jurisdictionPatentLink := make([]model.PatentSimpleFamilyJurisdictionLink, 0)

	minPriorityDateStr := minPriorityDate.Format("2006-01-02") // Adjust the format as needed

	claimsList, ok := data["priorityClaims"].([]interface{})
	if ok {
		for _, claim := range claimsList {
			if parsedClaim, ok := claim.(map[string]interface{}); ok {
				if claimDate, ok := parsedClaim["documentDate"].(string); ok && claimDate == minPriorityDateStr {
					if country, ok := parsedClaim["country"].(string); ok {
						jurisdictions = append(jurisdictions, model.SimpleFamilyJurisdiction{Name: country})
						jurisdictionPatentLink = append(jurisdictionPatentLink, model.PatentSimpleFamilyJurisdictionLink{
							JurisdictionName: country, PatentID: patentId})
					}
				}
			}
		}
	}
	return &jurisdictions, &jurisdictionPatentLink
}

type claimMapKey struct {
	claimID     string
	claimNumber int
}

type claimMapValue struct {
	dependent        []string
	dependentNumbers map[claimMapKey]struct{}
	claim            string
}

func (c *APIClient) parseClaim(data map[string]interface{}, patentId uuid.UUID) *[]model.Claim {
	claimsIds := make(map[string]struct{})
	claimsMap := make(map[claimMapKey]claimMapValue)
	claimsList, ok := data["claimsXml"].([]interface{})
	if ok {
		for _, value := range claimsList {
			if claimObject, ok := value.(map[string]interface{}); ok {
				singleClaim, ok := claimObject["xmlText"].(string)
				if !ok {
					continue
				}
				singleClaim = utils.RemoveHTMLTags(singleClaim)
				claimId, _ := claimObject["claimId"].(string)
				pattern := regexp.MustCompile(`\d{1,2}$`)
				match := pattern.FindString(claimId)
				var claimNumber int
				var err error
				if match != "" {
					claimNumber, err = strconv.Atoi(match)
					if err == nil {
						continue
					}
				}
				claimStruct := claimMapKey{claimID: claimId, claimNumber: claimNumber}

				if _, exists := claimsIds[claimId]; !exists {
					isDependent, ok := claimObject["isDependent"].(bool)
					if !ok {
						continue
					}
					if !isDependent {
						if _, exists := claimsMap[claimStruct]; !exists {
							claimsMap[claimStruct] = claimMapValue{claim: singleClaim + "\n", dependent: make([]string, 0),
								dependentNumbers: make(map[claimMapKey]struct{})}
						}
					} else {
						claimReferences, ok := claimObject["claimReferences"].([]string)
						if ok {
							if key, exists := checkClaimMapKey(claimReferences[0], claimsMap); exists {
								value, _ := claimsMap[key]
								value.dependentNumbers[claimStruct] = struct{}{}
								value.dependent = append(value.dependent, singleClaim)
								claimsMap[key] = value
							} else {
								for independentClaimNumber, dependentClaims := range claimsMap {
									if _, exists := checkDependentClaimMapKey(claimReferences[0], dependentClaims.dependentNumbers); exists {
										dependentClaims.dependent = append(dependentClaims.dependent, singleClaim)
										dependentClaims.dependentNumbers[claimStruct] = struct{}{}
										claimsMap[independentClaimNumber] = dependentClaims
									}
								}
							}
						}

					}
				} else {
					if key, exists := checkClaimMapKey(claimId, claimsMap); exists {
						value, _ := claimsMap[key]
						value.claim += singleClaim + "\n"
						claimsMap[key] = value
					} else {
						claimReferences, ok := claimObject["claimReferences"].([]string)
						if ok {
							if key, exists := checkClaimMapKey(claimReferences[0], claimsMap); exists {
								value := claimsMap[key]
								value.dependentNumbers[claimStruct] = struct{}{}
								value.dependent = append(value.dependent, singleClaim)
								claimsMap[key] = value
							} else {
								for independentClaimNumber, dependentClaim := range claimsMap {
									if _, exists := checkDependentClaimMapKey(claimReferences[0], dependentClaim.dependentNumbers); exists {
										value := claimsMap[independentClaimNumber]
										value.dependentNumbers[claimStruct] = struct{}{}
										value.dependent = append(value.dependent, singleClaim)
										claimsMap[independentClaimNumber] = value
									}
								}
							}
						} else {
							for independentClaimNumber, dependentClaim := range claimsMap {
								if _, exists := checkDependentClaimMapKey(claimId, dependentClaim.dependentNumbers); exists {
									value := claimsMap[independentClaimNumber]
									value.dependent = append(value.dependent, singleClaim)
									claimsMap[independentClaimNumber] = value
								}
							}
						}
					}
				}
			}
		}
	}
	claims := make([]model.Claim, 0, len(claimsMap))

	for key, value := range claimsMap {
		claims = append(claims, model.Claim{
			PatentID:         patentId,
			ClaimNumber:      key.claimNumber,
			IndependentClaim: value.claim,
			DependantClaims:  value.dependent,
		})
	}
	return &claims
}

func checkClaimMapKey(claimId string, claimMap map[claimMapKey]claimMapValue) (claimMapKey, bool) {
	for key, _ := range claimMap {
		if key.claimID == claimId {
			return key, true
		}
	}
	return claimMapKey{}, false
}

func checkDependentClaimMapKey(claimId string, claimMap map[claimMapKey]struct{}) (claimMapKey, bool) {
	for key, _ := range claimMap {
		if key.claimID == claimId {
			return key, true
		}
	}
	return claimMapKey{}, false
}
