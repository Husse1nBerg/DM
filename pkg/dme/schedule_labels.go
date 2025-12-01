package dme

// ScheduleLabel represents a schedule label with color and description
type ScheduleLabel struct {
	ID          int    `json:"id"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

