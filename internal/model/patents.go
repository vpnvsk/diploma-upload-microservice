package model

import (
	"github.com/google/uuid"
	"time"
)

// ParsedPatentsData holds the final parsed patents data
type ParsedPatentsData struct {
	Inventors               *[]Inventor                              `json:"inventors"`
	Assignees               *[]StandardizedCurrentAssignee           `json:"assignees"`
	Jurisdictions           *[]SimpleFamilyJurisdiction              `json:"jurisdictions"`
	Patent                  *ParsedPatent                            `json:"patent"`
	Claims                  *[]Claim                                 `json:"claims"`
	PatentInventorsLink     *[]PatentInventorLink                    `json:"patent_inventors_link"`
	PatentAssigneesLink     *[]PatentStandardizedCurrentAssigneeLink `json:"patent_assignees_link"`
	PatentJurisdictionsLink *[]PatentSimpleFamilyJurisdictionLink    `json:"patent_jurisdictions_link"`
}

type ParsedPatent struct {
	Id                             uuid.UUID `json:"id"`
	Title                          string    `json:"title"`
	Abstract                       string    `json:"abstract"`
	CPC                            string    `json:"cpc"`
	EarliestPriorityDate           time.Time `json:"earliest_priority_date"`
	EstimatedExpiryDate            time.Time `json:"estimated_expiry_date"`
	PublicationNumber              string    `json:"documentNumber"`
	CountOfCitedByPatents          int       `json:"count_of_cited_by_patents"`
	Description                    string    `json:"description"`
	BriefDescriptionOfDrawings     string    `json:"brief_description_of_drawings"`
	SimpleLegalStatus              string    `json:"simple_legal_status"`
	InpadocFamily                  string    `json:"inpadoc_family"`
	InpadocFamilyApplicationCount  int       `json:"inpadoc_family_application_count"`
	InpadocFamilyJurisdiction      string    `json:"inpadoc_family_jurisdiction"`
	InpadocFamilyJurisdictionCount int       `json:"inpadoc_family_jurisdiction_count"`
	Authority                      string    `json:"authority"`
	ApplicationDate                time.Time `json:"application_date"`
	ApplicationNumber              string    `json:"application_number"`
	IssueDate                      time.Time `json:"issue_date"`
	PublicationDate                time.Time `json:"publication_date"`
	FirstClaim                     string    `json:"first_claim"`
	TotalNumberOfClaims            int       `json:"total_number_of_claims"`
	TotalNumberOfIndependentClaims int       `json:"total_number_of_independent_claims"`
	FileURL                        string    `json:"file_url"`
}

type Inventor struct {
	FullName string `json:"full_name"`
}

type StandardizedCurrentAssignee struct {
	Name string `json:"name"`
}

type SimpleFamilyJurisdiction struct {
	Name string `json:"name"`
}

type Claim struct {
	PatentID         uuid.UUID `json:"patent_id"`
	ClaimNumber      int
	IndependentClaim string `json:"independent_claim"`
	DependantClaims  []string
}

type Patent struct {
	ID                             string `json:"id"`
	FirstClaim                     string `json:"first_claim"`
	TotalNumberOfClaims            int    `json:"total_number_of_claims"`
	TotalNumberOfIndependentClaims int    `json:"total_number_of_independent_claims"`
}

type PatentInventorLink struct {
	PatentID     uuid.UUID `json:"patent_id"`
	InventorName string    `json:"inventor_name"`
}

type PatentStandardizedCurrentAssigneeLink struct {
	PatentID     uuid.UUID `json:"patent_id"`
	AssigneeName string    `json:"assignee_name"`
}

type PatentSimpleFamilyJurisdictionLink struct {
	PatentID         uuid.UUID `json:"patent_id"`
	JurisdictionName string    `json:"jurisdiction_name"`
}

type ParsedPatentDataDB struct {
	Patents                 []*ParsedPatent
	Inventors               *[]Inventor
	Assignees               *[]StandardizedCurrentAssignee
	Jurisdictions           *[]SimpleFamilyJurisdiction
	Claims                  *[]Claim
	PatentInventorsLink     *[]PatentInventorLink
	PatentAssigneesLink     *[]PatentStandardizedCurrentAssigneeLink
	PatentJurisdictionsLink *[]PatentSimpleFamilyJurisdictionLink
}

type FilteredPatent struct {
	Title                    string     `json:"title"`
	PublicationNumber        string     `json:"publication_number"`
	EarliestPriorityDate     *time.Time `json:"earliest_priority_date"`
	EstimatedExpiryDate      *time.Time `json:"estimated_expiry_date"`
	InventorsNames           []string   `json:"inventors_names"`
	Assignee                 []string   `json:"assignee"`
	SimpleFamilyJurisdiction []string   `json:"simple_family_jurisdiction"`
	ApplicationDate          *time.Time `json:"application_date"`
	SimpleLegalStatus        *string    `json:"simple_legal_status"`
}

type FilteredPatentsResponse struct {
	Patents      *[]FilteredPatent       `json:"patents"`
	Statistics   *map[string]interface{} `json:"statistics"`
	TotalPatents int                     `json:"total_patents"`
}
