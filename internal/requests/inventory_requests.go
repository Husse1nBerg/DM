package requests

// RetrieveFuelRequest represents a request to retrieve fuel inventory
type RetrieveFuelRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
	FuelID   string `json:"fuelId" query:"fuelId"` // Optional - if empty, retrieves all
}

// RetrieveOnlineBillcodeListRequest represents a request to retrieve online billcodes
type RetrieveOnlineBillcodeListRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
}

// RetrieveOnlinePartsListRequest represents a request to retrieve online parts list
type RetrieveOnlinePartsListRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
}

// RetrieveQtyInfoRequest represents a request to retrieve quantity information for a part
type RetrieveQtyInfoRequest struct {
	SystemID     string `json:"systemId" query:"systemId" validate:"required"`
	PartNumber   string `json:"partNumber" query:"partNumber" validate:"required"`
	LocationCode string `json:"locationCode" query:"locationCode" validate:"required"`
}

// RetrievePartsKitRequest represents a request to retrieve a parts kit
type RetrievePartsKitRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
	KitID    string `json:"kitId" query:"kitId" validate:"required"`
}

// ListPartsKitsRequest represents a request to list all parts kits
type ListPartsKitsRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
}

// RetrievePurchaseOrderRequest represents a request to retrieve a purchase order
type RetrievePurchaseOrderRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
	POID     string `json:"poId" query:"poId" validate:"required"`
}

// ListPurchaseOrdersRequest represents a request to list purchase orders by date
type ListPurchaseOrdersRequest struct {
	SystemID  string `json:"systemId" query:"systemId" validate:"required"`
	StartDate string `json:"startDate" query:"startDate" validate:"required"` // Format: YYYY-MM-DD
	EndDate   string `json:"endDate" query:"endDate"`                         // Optional - for date range
}

// RetrieveSpecialOrderRequest represents a request to retrieve a special order
type RetrieveSpecialOrderRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
	OrderID  string `json:"orderId" query:"orderId" validate:"required"`
}

// ListCustomerSpecialOrdersRequest represents a request to list special orders for a customer
type ListCustomerSpecialOrdersRequest struct {
	SystemID   string `json:"systemId" query:"systemId" validate:"required"`
	CustomerID string `json:"customerId" query:"customerId" validate:"required"`
}

// ListReceivedSpecialOrdersRequest represents a request to list received special orders
type ListReceivedSpecialOrdersRequest struct {
	SystemID string `json:"systemId" query:"systemId" validate:"required"`
}

// SearchInventoryRequest represents a request to search inventory
type SearchInventoryRequest struct {
	SystemID   string `json:"systemId" query:"systemId" validate:"required"`
	SearchTerm string `json:"searchTerm" query:"searchTerm" validate:"required"`
}

// FindPartsRequest represents a request to find parts by part numbers
type FindPartsRequest struct {
	SystemID    string   `json:"systemId" validate:"required"`
	PartNumbers []string `json:"partNumbers" validate:"required,min=1,dive,required"`
}

// RetrieveInventoryRequest represents a request to retrieve inventory records
type RetrieveInventoryRequest struct {
	SystemID    string   `json:"systemId" validate:"required"`
	PartNumbers []string `json:"partNumbers" validate:"required,min=1,dive,required"`
}
