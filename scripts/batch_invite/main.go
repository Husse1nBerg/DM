package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	baseURL       = "https://dmwebapi.dockmaster.com/api/v1"
	defaultRoleID = "07a3c8fe-ddd4-4a3e-947f-0ab055f1bd79"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImY5NTA1ZjBlLWI5OWEtNGVhNC1hMDE4LWM1YzZiNDIwYWQ2ZCIsIm9yZ2FuaXphdGlvbklkIjoiNmY2OTMzMDYtODVmOC00Zjg3LTljOGItNzA0MGU0MzFlN2JjIiwibWFyaW5hSWQiOiJmZmQxNjYwMi1kZTI3LTQwNDItYTQzMC04NjRjYzVhZmYxZTAiLCJuYW1lIjoiQW5kcmV3IFNhbWVoIiwiZW1haWwiOiJhbmRyZXcuc2FtZWhAZG9ja21hc3Rlci5jb20iLCJyb2xlSWQiOiJjYmEwZmMzOC0yNjBjLTQyMGUtOTQ0NC02YjJlYzYxMGVhYmUiLCJleHAiOjE3NTAyNzIxNDF9.DKM9vUc3c8SlAZyk04eGkUIvYu1vfPe-5-sAujcSMfQ"

type OnboardingData struct {
	Name        string
	Email       string
	APIEmail    string
	APIPassword string
	SystemID    string
}

type User struct {
	FirstName   string
	LastName    string
	Email       string
	CompanyName string
}

type MarinaByEmailResponse struct {
	Data struct {
		ID             string `json:"id"`
		OrganizationID string `json:"organizationId"`
		Name           string `json:"name"`
		Email          string `json:"email"`
	} `json:"data"`
}

type UserInviteRequest struct {
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	MarinaID       string `json:"marinaId"`
	OrganizationID string `json:"organizationId"`
	RoleID         string `json:"roleId"`
}

func main() {
	log.Println("Starting batch invite process...")

	// Read onboarding data
	onboardingData, err := readOnboardingData("scripts/batch_invite/onboarding_data.csv")
	if err != nil {
		log.Fatalf("Failed to read onboarding data: %v", err)
	}
	log.Printf("Loaded %d marina records from onboarding data", len(onboardingData))

	// Read users list
	users, err := readUsersList("scripts/batch_invite/users-list.csv")
	if err != nil {
		log.Fatalf("Failed to read users list: %v", err)
	}
	log.Printf("Loaded %d users from users list", len(users))

	// Create marina name to email mapping
	marinaEmailMap := make(map[string]string)
	for _, marina := range onboardingData {
		marinaEmailMap[strings.ToLower(marina.Name)] = marina.Email
	}

	// Process each user
	processedMarinas := make(map[string]bool)
	invitesSent := 0
	invitesFailed := 0
	alreadyExists := 0

	for _, user := range users {
		log.Printf("Processing user: %s %s (%s) - Company: %s", user.FirstName, user.LastName, user.Email, user.CompanyName)

		// Find matching marina email
		marinaEmail, found := marinaEmailMap[strings.ToLower(user.CompanyName)]
		if !found {
			log.Printf("Warning: No marina found for company: %s", user.CompanyName)
			continue
		}

		// Get marina details by email
		marinaDetails, err := getMarinaByEmail(marinaEmail, token)
		if err != nil {
			log.Printf("Error: Failed to get marina details for email %s: %v", marinaEmail, err)
			continue
		}

		log.Printf("Found marina: %s (ID: %s, OrgID: %s)", marinaDetails.Data.Name, marinaDetails.Data.ID, marinaDetails.Data.OrganizationID)

		// Send invitation
		result, err := sendUserInvitation(user, marinaDetails.Data.ID, marinaDetails.Data.OrganizationID, token)
		if err != nil {
			if result == "already_exists" {
				log.Printf("User %s already exists - skipping", user.Email)
				alreadyExists++
			} else {
				log.Printf("Error: Failed to send invitation to %s: %v", user.Email, err)
				invitesFailed++
			}
			continue
		}

		log.Printf("Successfully sent invitation to %s", user.Email)
		invitesSent++

		// Track processed marinas for summary
		processedMarinas[marinaDetails.Data.Name] = true
	}

	// Print summary
	log.Printf("Batch invite process completed!")
	log.Printf("Total invitations sent: %d", invitesSent)
	log.Printf("Total users already exist: %d", alreadyExists)
	log.Printf("Total invitations failed: %d", invitesFailed)
	log.Printf("Total marinas processed: %d", len(processedMarinas))
}

func readOnboardingData(filePath string) ([]OnboardingData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	var data []OnboardingData
	// Skip header row
	for i, record := range records[1:] {
		if len(record) < 5 {
			log.Printf("Row %d: insufficient columns, skipping", i+2)
			continue
		}

		entry := OnboardingData{
			Name:        strings.TrimSpace(record[0]),
			Email:       strings.ToLower(strings.TrimSpace(record[1])),
			APIEmail:    strings.TrimSpace(record[2]),
			APIPassword: strings.TrimSpace(record[3]),
			SystemID:    strings.TrimSpace(record[4]),
		}
		data = append(data, entry)
	}

	return data, nil
}

func readUsersList(filePath string) ([]User, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	var users []User
	// Skip header row
	for i, record := range records[1:] {
		if len(record) < 4 {
			log.Printf("Row %d: insufficient columns, skipping", i+2)
			continue
		}

		user := User{
			FirstName:   strings.TrimSpace(record[0]),
			LastName:    strings.TrimSpace(record[1]),
			Email:       strings.ToLower(strings.TrimSpace(record[2])),
			CompanyName: strings.TrimSpace(record[3]),
		}
		users = append(users, user)
	}

	return users, nil
}

func getMarinaByEmail(email, authToken string) (*MarinaByEmailResponse, error) {
	url := fmt.Sprintf("%s/marinas/by-email?email=%s", baseURL, email)

	resp, err := makeAPIRequest("GET", url, nil, authToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var marinaResp MarinaByEmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&marinaResp); err != nil {
		return nil, err
	}

	return &marinaResp, nil
}

func sendUserInvitation(user User, marinaID, organizationID, authToken string) (string, error) {
	reqBody := UserInviteRequest{
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		MarinaID:       marinaID,
		OrganizationID: organizationID,
		RoleID:         defaultRoleID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := makeAPIRequest("POST", baseURL+"/user/invite", jsonBody, authToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Check if it's the "Email already taken" error
		if resp.StatusCode == http.StatusBadRequest && strings.Contains(string(body), "Email already taken") {
			return "already_exists", fmt.Errorf("user already exists")
		}

		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return "success", nil
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
