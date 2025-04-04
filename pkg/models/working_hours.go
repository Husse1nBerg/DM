package models

import (
	"encoding/json"
	"fmt"
)

// WorkingHours represents the operating hours for each day of the week for a marina
type WorkingHours struct {
	Monday    string `json:"monday" example:"9:00 AM - 5:00 PM"`
	Tuesday   string `json:"tuesday" example:"9:00 AM - 5:00 PM"`
	Wednesday string `json:"wednesday" example:"9:00 AM - 5:00 PM"`
	Thursday  string `json:"thursday" example:"9:00 AM - 5:00 PM"`
	Friday    string `json:"friday" example:"9:00 AM - 5:00 PM"`
	Saturday  string `json:"saturday" example:"9:00 AM - 5:00 PM"`
	Sunday    string `json:"sunday" example:"9:00 AM - 5:00 PM"`
}

// ToBytes converts the WorkingHours struct to a JSON byte array for database storage
func (wh *WorkingHours) ToBytes() ([]byte, error) {
	return json.Marshal(wh)
}

// FromBytes populates the WorkingHours struct from a JSON byte array
func (wh *WorkingHours) FromBytes(data []byte) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, wh)
}

// String returns a string representation of the WorkingHours for debugging
func (wh *WorkingHours) String() string {
	data, err := json.Marshal(wh)
	if err != nil {
		return fmt.Sprintf("Error marshaling WorkingHours: %v", err)
	}
	return string(data)
}

// DefaultWorkingHours returns a WorkingHours struct with default 9-5 working hours
func DefaultWorkingHours() *WorkingHours {
	return &WorkingHours{
		Monday:    "9:00 AM - 5:00 PM",
		Tuesday:   "9:00 AM - 5:00 PM",
		Wednesday: "9:00 AM - 5:00 PM",
		Thursday:  "9:00 AM - 5:00 PM",
		Friday:    "9:00 AM - 5:00 PM",
		Saturday:  "9:00 AM - 5:00 PM",
		Sunday:    "9:00 AM - 5:00 PM",
	}
}
