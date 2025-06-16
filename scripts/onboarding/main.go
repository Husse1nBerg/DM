package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	baseURL       = "https://dmwebapi.dockmaster.com/api/v1"
	notesPlanID   = "eb6d21c2-363b-4c33-847d-7a394b70f19c"
	storagePlanID = "fc74be78-bbe2-432f-8544-d390706f19b8"
	defaultRoleID = "07a3c8fe-ddd4-4a3e-947f-0ab055f1bd79"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImY5NTA1ZjBlLWI5OWEtNGVhNC1hMDE4LWM1YzZiNDIwYWQ2ZCIsIm9yZ2FuaXphdGlvbklkIjoiNmY2OTMzMDYtODVmOC00Zjg3LTljOGItNzA0MGU0MzFlN2JjIiwibWFyaW5hSWQiOiJmZmQxNjYwMi1kZTI3LTQwNDItYTQzMC04NjRjYzVhZmYxZTAiLCJuYW1lIjoiQW5kcmV3IFNhbWVoIiwiZW1haWwiOiJhbmRyZXcuc2FtZWhAZG9ja21hc3Rlci5jb20iLCJyb2xlSWQiOiJjYmEwZmMzOC0yNjBjLTQyMGUtOTQ0NC02YjJlYzYxMGVhYmUiLCJleHAiOjE3NTAxMTk2MTV9.UqvY-zbBUetZJ0cwGwtjEe2kOVoCMdT-1E2VQQUqoBA"

type OnboardingData struct {
	Name        string
	Email       string
	APIEmail    string
	APIPassword string
	SystemID    string
}

type Address struct {
	City       string  `json:"city"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	PostalCode string  `json:"postal_code"`
	State      string  `json:"state"`
	Street     string  `json:"street"`
}

type WorkingHours struct {
	Monday    string `json:"monday"`
	Tuesday   string `json:"tuesday"`
	Wednesday string `json:"wednesday"`
	Thursday  string `json:"thursday"`
	Friday    string `json:"friday"`
	Saturday  string `json:"saturday"`
	Sunday    string `json:"sunday"`
}

type OrganizationRequest struct {
	Email    string  `json:"email"`
	Name     string  `json:"name"`
	Address  Address `json:"address"`
	Country  string  `json:"country"`
	IsActive bool    `json:"is_active"`
}

type MarinaRequest struct {
	Email               string       `json:"email"`
	Name                string       `json:"name"`
	OrganizationID      string       `json:"organizationId"`
	Address             Address      `json:"address"`
	Country             string       `json:"country"`
	WorkingHours        WorkingHours `json:"workingHours"`
	MaxUsers            int          `json:"maxUsers"`
	IsActive            bool         `json:"isActive"`
	IsTest              bool         `json:"isTest"`
	NotesMessagesPlanID string       `json:"notesMessagesPlanId"`
	StoragePlanID       string       `json:"storagePlanId"`
}

type DMECredentialsRequest struct {
	OrganizationID string `json:"organizationId"`
	Username       string `json:"username"`
	IsOldAPI       bool   `json:"isOldApi"`
	Password       string `json:"password"`
}

type SystemIDLinkRequest struct {
	MarinaID string `json:"marinaId"`
}

type UserInviteRequest struct {
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	MarinaID       string `json:"marinaId"`
	OrganizationID string `json:"organizationId"`
	RoleID         string `json:"roleId"`
}

type OrganizationResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type MarinaResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type SystemIDEntry struct {
	ID       string `json:"id"`
	SystemID string `json:"systemId"`
}

type SystemIDsResponse struct {
	Data []SystemIDEntry `json:"data"`
}

func main() {
	var (
		name        = flag.String("name", "", "Organization/Marina name")
		email       = flag.String("email", "", "Email address")
		apiEmail    = flag.String("api-email", "", "DME API email")
		apiPassword = flag.String("api-password", "", "DME API password")
		systemID    = flag.String("system-id", "", "System ID")
		authToken   = flag.String("auth-token", token, "Authorization bearer token")
		csvFile     = flag.String("csv", "scripts/onboarding/onboarding_data.csv", "CSV file path")
		useCsv      = flag.Bool("use-csv", false, "Use CSV file instead of individual flags")
	)
	flag.Parse()

	if *useCsv {
		processCsvFile(*csvFile, *authToken)
	} else {
		if *name == "" || *email == "" || *apiEmail == "" || *apiPassword == "" || *systemID == "" || *authToken == "" {
			log.Fatal("All parameters are required: -name, -email, -api-email, -api-password, -system-id, -auth-token")
		}

		data := OnboardingData{
			Name:        *name,
			Email:       *email,
			APIEmail:    *apiEmail,
			APIPassword: *apiPassword,
			SystemID:    *systemID,
		}

		err := processOnboarding(data, *authToken)
		if err != nil {
			log.Fatalf("Onboarding failed: %v", err)
		}
	}
}

func processCsvFile(csvPath, authToken string) {
	if authToken == "" {
		log.Fatal("Authorization token is required")
	}

	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	if len(records) < 2 {
		log.Fatal("CSV file must have at least a header and one data row")
	}

	// Skip header row
	for i, record := range records[1:] {
		if len(record) < 5 {
			log.Printf("Row %d: insufficient columns, skipping", i+2)
			continue
		}

		data := OnboardingData{
			Name:        strings.TrimSpace(record[0]),
			Email:       strings.TrimSpace(record[1]),
			APIEmail:    strings.TrimSpace(record[2]),
			APIPassword: strings.TrimSpace(record[3]),
			SystemID:    strings.TrimSpace(record[4]),
		}

		log.Printf("Processing row %d: %s (%s)", i+2, data.Name, data.Email)

		err := processOnboarding(data, authToken)
		if err != nil {
			log.Printf("Row %d failed: %v", i+2, err)
			continue
		}

		log.Printf("Row %d completed successfully", i+2)
	}
}

func processOnboarding(data OnboardingData, authToken string) error {
	log.Printf("Starting onboarding for: %s (%s)", data.Name, data.Email)

	// Step 1: Create Organization
	log.Println("Step 1: Creating organization...")
	orgID, err := createOrganization(data, authToken)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	log.Printf("Organization created with ID: %s", orgID)

	// Step 2: Create Marina
	log.Println("Step 2: Creating marina...")
	marinaID, err := createMarina(data, orgID, authToken)
	if err != nil {
		return fmt.Errorf("failed to create marina: %w", err)
	}
	log.Printf("Marina created with ID: %s", marinaID)

	// Step 3: Create DME Credentials
	log.Println("Step 3: Creating DME credentials...")
	err = createDMECredentials(data, orgID, authToken)
	if err != nil {
		return fmt.Errorf("failed to create DME credentials: %w", err)
	}
	log.Println("DME credentials created successfully")

	// Step 4: Get System IDs and find the matching one
	log.Println("Step 4: Finding system ID...")
	systemIDObjectID, err := findSystemIDObjectID(data.SystemID, orgID, authToken)
	if err != nil {
		return fmt.Errorf("failed to find system ID: %w", err)
	}
	log.Printf("Found system ID object with ID: %s", systemIDObjectID)

	// Step 5: Link System ID to Marina
	log.Println("Step 5: Linking system ID to marina...")
	err = linkSystemIDToMarina(systemIDObjectID, marinaID, authToken)
	if err != nil {
		return fmt.Errorf("failed to link system ID to marina: %w", err)
	}
	log.Println("System ID linked to marina successfully")

	// Step 6: Send User Invitation
	log.Println("Step 6: Sending user invitation...")
	err = sendUserInvitation(data, orgID, marinaID, authToken)
	if err != nil {
		return fmt.Errorf("failed to send user invitation: %w", err)
	}
	log.Println("User invitation sent successfully")

	log.Printf("Onboarding completed for: %s (%s)", data.Name, data.Email)
	return nil
}

func createOrganization(data OnboardingData, authToken string) (string, error) {
	reqBody := OrganizationRequest{
		Email: data.Email,
		Name:  data.Name,
		Address: Address{
			City:       "",
			Country:    "USA",
			Latitude:   1.03,
			Longitude:  2.36,
			PostalCode: "",
			State:      "",
			Street:     "",
		},
		Country:  "USA",
		IsActive: true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := makeAPIRequest("POST", baseURL+"/organizations", jsonBody, authToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var orgResp OrganizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&orgResp); err != nil {
		return "", err
	}

	return orgResp.Data.ID, nil
}

func createMarina(data OnboardingData, orgID, authToken string) (string, error) {
	reqBody := MarinaRequest{
		Email:          data.Email,
		Name:           data.Name,
		OrganizationID: orgID,
		Address: Address{
			City:       "",
			Country:    "USA",
			Latitude:   1.03,
			Longitude:  2.36,
			PostalCode: "",
			State:      "",
			Street:     "",
		},
		Country: "USA",
		WorkingHours: WorkingHours{
			Monday:    "9:00 AM - 5:00 PM",
			Tuesday:   "9:00 AM - 5:00 PM",
			Wednesday: "9:00 AM - 5:00 PM",
			Thursday:  "9:00 AM - 5:00 PM",
			Friday:    "9:00 AM - 5:00 PM",
			Saturday:  "9:00 AM - 5:00 PM",
			Sunday:    "9:00 AM - 5:00 PM",
		},
		MaxUsers:            100,
		IsActive:            true,
		IsTest:              false,
		NotesMessagesPlanID: notesPlanID,
		StoragePlanID:       storagePlanID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := makeAPIRequest("POST", baseURL+"/marinas", jsonBody, authToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var marinaResp MarinaResponse
	if err := json.NewDecoder(resp.Body).Decode(&marinaResp); err != nil {
		return "", err
	}

	return marinaResp.Data.ID, nil
}

func createDMECredentials(data OnboardingData, orgID, authToken string) error {
	reqBody := DMECredentialsRequest{
		OrganizationID: orgID,
		Username:       data.APIEmail,
		IsOldAPI:       false,
		Password:       data.APIPassword,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := makeAPIRequest("POST", baseURL+"/dme/credentials", jsonBody, authToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func findSystemIDObjectID(systemID, orgID, authToken string) (string, error) {
	url := fmt.Sprintf("%s/dme/sysids/organization/%s", baseURL, orgID)

	resp, err := makeAPIRequest("GET", url, nil, authToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var sysIDsResp SystemIDsResponse
	if err := json.NewDecoder(resp.Body).Decode(&sysIDsResp); err != nil {
		return "", err
	}

	for _, entry := range sysIDsResp.Data {
		if entry.SystemID == systemID {
			return entry.ID, nil
		}
	}

	return "", fmt.Errorf("system ID %s not found in organization %s", systemID, orgID)
}

func linkSystemIDToMarina(systemIDObjectID, marinaID, authToken string) error {
	reqBody := SystemIDLinkRequest{
		MarinaID: marinaID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/dme/sysids/%s/link", baseURL, systemIDObjectID)

	resp, err := makeAPIRequest("PATCH", url, jsonBody, authToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func sendUserInvitation(data OnboardingData, orgID, marinaID, authToken string) error {
	fmt.Println("Sending user invitation for: not ", data.Email)
	reqBody := UserInviteRequest{
		Email: "andrey.safonov+bbw@aspiresoftware.com",
		// Email:          data.Email,
		FirstName:      "Admin",
		LastName:       "User",
		MarinaID:       marinaID,
		OrganizationID: orgID,
		RoleID:         defaultRoleID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := makeAPIRequest("POST", baseURL+"/user/invite", jsonBody, authToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func makeAPIRequest(method, url string, body []byte, authToken string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	client := &http.Client{}
	return client.Do(req)
}
