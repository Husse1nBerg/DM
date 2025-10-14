package responses

// ScheduleResponse represents a schedule response with appointments
type ScheduleResponse struct {
	Data interface{} `json:"data"`
}

// ScheduleUpdateResponse represents the response from a schedule update
type ScheduleUpdateResponse struct {
	Data interface{} `json:"data"`
}

// ConvertScheduleResponse converts interface{} to ScheduleResponse
func ConvertScheduleResponse(dmeResponse *interface{}) *ScheduleResponse {
	return &ScheduleResponse{
		Data: *dmeResponse,
	}
}

// ConvertScheduleUpdateResponse converts interface{} to ScheduleUpdateResponse
func ConvertScheduleUpdateResponse(dmeResponse *interface{}) *ScheduleUpdateResponse {
	return &ScheduleUpdateResponse{
		Data: *dmeResponse,
	}
}

