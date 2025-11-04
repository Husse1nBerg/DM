package requests

// ScheduleRetrieveRequest represents a request to retrieve schedule
type ScheduleRetrieveRequest struct {
	LocationCode string `query:"locationCode" validate:"required"`
	StartDate    string `query:"startDate" validate:"required"`
	SessionID    string `query:"sessionId"` // Optional - empty to retrieve all appointments
}

// ScheduleRetrieveForManagerRequest represents a request to retrieve schedule for a manager
type ScheduleRetrieveForManagerRequest struct {
	LocationCode string `query:"locationCode" validate:"required"`
	StartDate    string `query:"startDate" validate:"required"`
	ManagerID    string `query:"managerId" validate:"required"`
	SessionID    string `query:"sessionId"` // Optional - empty to retrieve all appointments
}

// ScheduleRetrieveForTechRequest represents a request to retrieve schedule for a technician
type ScheduleRetrieveForTechRequest struct {
	LocationCode string `query:"locationCode" validate:"required"`
	StartDate    string `query:"startDate" validate:"required"`
	TechID       string `query:"techId" validate:"required"`
	SessionID    string `query:"sessionId"` // Optional - empty to retrieve all appointments
}

// ScheduleRetrieveForWorkOrderRequest represents a request to retrieve schedule for a work order
type ScheduleRetrieveForWorkOrderRequest struct {
	LocationCode string `query:"locationCode" validate:"required"`
	StartDate    string `query:"startDate" validate:"required"`
	WorkOrderID  string `query:"workOrderId" validate:"required"`
	SessionID    string `query:"sessionId"` // Optional - empty to retrieve all appointments
}

// ScheduleWorkOrderScheduleRequest represents a request to retrieve work order schedule
type ScheduleWorkOrderScheduleRequest struct {
	LocationCode string `query:"locationCode" validate:"required"`
	WorkOrderID  string `query:"workOrderId" validate:"required"`
	SessionID    string `query:"sessionId"` // Optional - empty to retrieve all appointments
}

// ScheduleOperationScheduleRequest represents a request to retrieve operation schedule
type ScheduleOperationScheduleRequest struct {
	LocationCode string `query:"locationCode" validate:"required"`
	WorkOrderID  string `query:"workOrderId" validate:"required"`
	Opcode       string `query:"opcode" validate:"required"`
	SessionID    string `query:"sessionId"` // Optional - empty to retrieve all appointments
}

// ScheduleAppointmentUpdate represents an appointment update item
type ScheduleAppointmentUpdate struct {
	ID              string  `json:"id"` // ID can be empty for new appointments - DME API will generate it
	TechID          string  `json:"techId" validate:"required"`
	WorkOrderID     string  `json:"workOrderId" validate:"required"`
	Opcode          string  `json:"opcode" validate:"required"`
	LocationCode    string  `json:"locationCode" validate:"required"`
	EstHours        float64 `json:"estHours"`
	ApptDate        string  `json:"apptDate" validate:"required"` // YYYY-MM-DD format
	StartTime       string  `json:"startTime" validate:"required"`
	EndTime         string  `json:"endTime" validate:"required"`
	Status          string  `json:"status"`
	Label           int     `json:"label"`
	ApptComplete    bool    `json:"apptComplete"`
	OpLaborFinished bool    `json:"opLaborFinished"`
	IsDeleted       bool    `json:"isDeleted"`
}

// ScheduleUpdateRequest represents a request to update schedule
type ScheduleUpdateRequest struct {
	LocationCode string                       `json:"locationCode" validate:"required"`
	ClerkID      string                       `json:"clerkId" validate:"required"`
	SessionID    string                       `json:"sessionId" validate:"required"`
	Appointments []ScheduleAppointmentUpdate  `json:"appointments" validate:"required,dive"`
}

// ScheduleResolveMergeConflictRequest represents a request to resolve merge conflicts
type ScheduleResolveMergeConflictRequest struct {
	LocationCode string                       `json:"locationCode" validate:"required"`
	ClerkID      string                       `json:"clerkId" validate:"required"`
	SessionID    string                       `json:"sessionId" validate:"required"`
	Appointments []ScheduleAppointmentUpdate  `json:"appointments" validate:"required,dive"`
}

