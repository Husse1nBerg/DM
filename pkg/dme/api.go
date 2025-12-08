package dme

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// -----
// General API
// -----

// ListLocations lists all locations
func (c *Client) ListLocations(ctx context.Context, organizationID uuid.UUID, systemID string) ([]Location, error) {
	var result []Location
	endpoint := "/General/Locations"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	return result, nil
}

// ListDepartments retrieves a list of departments
func (c *Client) ListDepartments(ctx context.Context, organizationID uuid.UUID, systemID string) ([]Department, error) {
	var result []Department
	endpoint := "/Departments/List"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	return result, nil
}

// RetrieveCustomerQuotes retrieves customer quotes for Unit Sales module
func (c *Client) RetrieveCustomerQuotes(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) ([]CustomerContract, error) {
	var result []CustomerContract
	endpoint := fmt.Sprintf("/UnitSales/RetrieveCustomerQuotes?CustomerId=%s", customerID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer quotes: %w", err)
	}

	return result, nil
}

// ListClerks retrieves a list of system clerks (users)
func (c *Client) ListClerks(ctx context.Context, organizationID uuid.UUID, systemID string, includeInactive bool) ([]Clerk, error) {
	var result []Clerk
	endpoint := "/General/Clerks/List"

	// Add query parameter for includeInactive
	params := map[string]string{
		"includeInactive": fmt.Sprintf("%t", includeInactive),
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list clerks: %w", err)
	}

	return result, nil
}

// RetrieveClerk retrieves a specific clerk (user) record
func (c *Client) RetrieveClerk(ctx context.Context, clerkID string, organizationID uuid.UUID, systemID string) (*Clerk, error) {
	var result Clerk
	endpoint := fmt.Sprintf("/General/Clerks?Id=%s", clerkID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve clerk: %w", err)
	}

	return &result, nil
}

// RetrieveScheduleLabels retrieves the list of schedule labels configured in the system
func (c *Client) RetrieveScheduleLabels(ctx context.Context, organizationID uuid.UUID, systemID string) ([]ScheduleLabel, error) {
	var result []ScheduleLabel
	endpoint := "/Service/Schedule/Labels"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schedule labels: %w", err)
	}

	return result, nil
}

// -----
// Customer API
// -----

// CustomerSearch searches for customers
func (c *Client) CustomerSearch(ctx context.Context, searchTerm string, directHit string, organizationID uuid.UUID, systemID string) (*[]CustomerSearch, error) {
	var result []CustomerSearch
	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credential: %w", err)
	}

	if credential.IsOldApi != nil && *credential.IsOldApi {
		payload := map[string]interface{}{
			"SearchString": searchTerm,
			"DirectHit":    directHit,
		}
		endpoint := "/Customers/Search"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search customers: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Customers/Search?SearchString=%s&DirectHit=%s", searchTerm, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search customers: %w", err)
		}
	}

	return &result, nil
}

// ListCustomers retrieves a full list of customers
func (c *Client) CustomersList(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) (*CustomerListMinimal, error) {
	var result CustomerListMinimal

	endpoint := fmt.Sprintf("/Customers/ListByPage?Page=%d&PageSize=%d", page, pageSize)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get full customers list: %w", err)
	}
	return &result, nil
}

// RetrieveCustomer retrieves customer information
func (c *Client) CustomerRetrieve(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (*Customer, error) {
	var result Customer

	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credential: %w", err)
	}

	var endpoint string
	if credential.IsOldApi != nil && *credential.IsOldApi {
		endpoint = fmt.Sprintf("/Customers/RetrieveCustomer/%s", customerID)
	} else {
		endpoint = fmt.Sprintf("/Customers/RetrieveCustomer?CustomerId=%s", customerID)
	}

	err = c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer information: %w", err)
	}

	return &result, nil
}

// CustomersListShort retrieves a short list of customers
func (c *Client) CustomersListShort(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) (*CustomerListShort, error) {
	var result CustomerListShort
	endpoint := fmt.Sprintf("/Customers/ListByPage?Page=%d&PageSize=%d", page, pageSize)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers by page: %w", err)
	}

	return &result, nil
}

// RetrieveCustomersFiltered retrieves a list of customers with optional filters
func (c *Client) RetrieveCustomersFiltered(ctx context.Context, lastModifiedDate string, emailAddress string, organizationID uuid.UUID, systemID string) ([]Customer, error) {
	var result []Customer

	// Build endpoint with optional query parameters
	endpoint := "/Customers/RetrieveCustomers?"
	params := []string{}

	if lastModifiedDate != "" {
		params = append(params, fmt.Sprintf("LastModifiedDate=%s", lastModifiedDate))
	}

	if emailAddress != "" {
		params = append(params, fmt.Sprintf("EmailAddress=%s", emailAddress))
	}

	// Join parameters with &
	if len(params) > 0 {
		endpoint += strings.Join(params, "&")
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customers with filters: %w", err)
	}

	return result, nil
}

// RetrieveCustomersPaginated retrieves customers with category codes in a paginated format
// Used for esignature and mass notification features
func (c *Client) RetrieveCustomersPaginated(ctx context.Context, page, pageSize int, listName, lastModifiedDate, emailAddress string, organizationID uuid.UUID, systemID string) (*CustomerWithCategoryCodesPage, error) {
	var result CustomerWithCategoryCodesPage

	// Build endpoint with query parameters
	endpoint := fmt.Sprintf("/Customers/RetrieveCustomersPaginated?Page=%d&PageSize=%d", page, pageSize)

	if listName != "" {
		endpoint += fmt.Sprintf("&ListName=%s", listName)
	}

	if lastModifiedDate != "" {
		endpoint += fmt.Sprintf("&LastModifiedDate=%s", lastModifiedDate)
	}

	if emailAddress != "" {
		endpoint += fmt.Sprintf("&EmailAddress=%s", emailAddress)
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customers paginated: %w", err)
	}

	return &result, nil
}

// RetrieveCustomerContracts retrieves customer contracts for Unit Sales module
func (c *Client) RetrieveCustomerContracts(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) ([]CustomerContract, error) {
	var result []CustomerContract
	endpoint := fmt.Sprintf("/UnitSales/RetrieveCustomerContracts?CustomerId=%s", customerID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer contracts: %w", err)
	}

	return result, nil
}

// UpdateCustomer updates a customer
func (c *Client) CustomerUpdate(ctx context.Context, payload interface{}, organizationID uuid.UUID, systemID string) (*Customer, error) {
	var result CustomerCreateUpdateResponse

	endpoint := "/DockMaster/Customers/UpdateCustomer"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		payload,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}
	updatedCustomer, err := c.CustomerRetrieve(ctx, result.CustomerID, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated customer: %w", err)
	}
	return updatedCustomer, nil
}

// CreateCustomer creates a new customer
func (c *Client) CreateCustomer(ctx context.Context, customer *CustomerCreate, organizationID uuid.UUID, systemID string) (*Customer, error) {
	var result CustomerCreateUpdateResponse
	endpoint := "/DockMaster/Customers/UpdateCustomer"

	// Initialize empty arrays if nil
	if customer.CustomInformation == nil {
		customer.CustomInformation = []CustomInformation{}
	}
	if customer.Attachments == nil {
		customer.Attachments = []Attachment{}
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		customer,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	createdCustomer, err := c.CustomerRetrieve(ctx, result.CustomerID, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created customer: %w", err)
	}
	return createdCustomer, nil
}

// -----
// Boat API
// -----

// BoatsList retrieves a list of boats
func (c *Client) BoatsList(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) (*BoatListMinimal, error) {
	var result BoatListMinimal
	endpoint := fmt.Sprintf("/Boats/ListByPage?Page=%d&PageSize=%d", page, pageSize)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list boats: %w", err)
	}

	return &result, nil
}

// RetrieveBoatByID retrieves a boat by its ID
func (c *Client) RetrieveBoatByID(ctx context.Context, boatID string, organizationID uuid.UUID, systemID string) (*Boat, error) {
	var result Boat
	endpoint := fmt.Sprintf("/Boats/RetrieveBoat?BoatId=%s", boatID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve boat by ID: %w", err)
	}

	return &result, nil
}

// RetrieveBoatsForCustomer retrieves boats associated with a customer
func (c *Client) RetrieveBoatsForCustomer(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) ([]Boat, error) {
	var result []Boat
	endpoint := fmt.Sprintf("/Boats/RetrieveBoats?CustomerId=%s", customerID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve boats for customer: %w", err)
	}

	return result, nil
}

// RetrieveBoatsWithInsurance retrieves boats associated with insurance
func (c *Client) RetrieveBoatsWithInsurance(ctx context.Context, organizationID uuid.UUID, systemID string) ([]Boat, error) {
	var result []Boat
	endpoint := fmt.Sprintf("/Boats/RetrieveBoats?HasInsurance=true")

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve boats with insurance: %w", err)
	}

	return result, nil
}

// SearchBoats searches for boats
func (c *Client) SearchBoats(ctx context.Context, searchTerm string, directHit bool, organizationID uuid.UUID, systemID string) ([]BoatSearch, error) {
	var result []BoatSearch
	endpoint := fmt.Sprintf("/Boats/Search?SearchString=%s&DirectHit=%t", searchTerm, directHit)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search boats: %w", err)
	}

	return result, nil
}

// BoatsListNewOrChanged retrieves boats created or changed after a specific date with pagination
func (c *Client) BoatsListNewOrChanged(ctx context.Context, lastUpdate string, page int, pageSize int, listName string, organizationID uuid.UUID, systemID string) (*BoatList, error) {
	var result BoatList

	// Build endpoint with query parameters
	endpoint := fmt.Sprintf("/Boats/ListNewOrChanged?LastUpdate=%s&Page=%d&PageSize=%d", lastUpdate, page, pageSize)
	if listName != "" {
		endpoint += fmt.Sprintf("&ListName=%s", listName)
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list new or changed boats: %w", err)
	}

	return &result, nil
}

// RetrieveBoatsFiltered retrieves boats with optional filters
func (c *Client) RetrieveBoatsFiltered(ctx context.Context, customerID string, lastUpdateDate string, hasInsurance bool, organizationID uuid.UUID, systemID string) ([]Boat, error) {
	var result []Boat

	// Build endpoint with optional query parameters
	endpoint := "/Boats/RetrieveBoats?"
	params := []string{}

	if customerID != "" {
		params = append(params, fmt.Sprintf("CustomerId=%s", customerID))
	}

	if lastUpdateDate != "" {
		params = append(params, fmt.Sprintf("LastUpdateDate=%s", lastUpdateDate))
	}

	if hasInsurance {
		params = append(params, "HasInsurance=true")
	}

	// Join parameters with &
	if len(params) > 0 {
		endpoint += strings.Join(params, "&")
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve boats with filters: %w", err)
	}

	return result, nil
}

// UpdateBoat updates a boat
func (c *Client) UpdateBoat(ctx context.Context, boat *BoatUpdate, organizationID uuid.UUID, systemID string) (*Boat, error) {
	var result BoatCreateUpdateResponse
	endpoint := "/DockMaster/Boats/UpdateBoat"
	//endpoint := "v1/Boats/UpdateBoat"

	// Initialize all array fields if they are null
	if boat.Motors == nil {
		boat.Motors = []Motor{}
	}
	// Ensure Drives and Generators are present to satisfy DME validation
	if boat.Drives == nil {
		boat.Drives = []Drive{}
	}
	if boat.Generators == nil {
		boat.Generators = []Generator{}
	}

	if boat.BoatDescriptionCodes == nil {
		boat.BoatDescriptionCodes = []BoatDescriptionCode{}
	}
	if boat.CustomInformation == nil {
		boat.CustomInformation = []CustomInformation{}
	}
	if boat.OperationsHistory == nil {
		boat.OperationsHistory = []OperationHistory{}
	}
	if boat.Attachments == nil {
		boat.Attachments = []Attachment{}
	}

	boat.BillingCodes = nil

	// Initialize Slip if it's nil
	if boat.Slip == (Slip{}) {
		boat.Slip = Slip{
			LastModifedDate: time.Now().Format("2006-01-02T15:04:05"),
		}
	}

	// Set LastModified if empty
	boat.LastModified = time.Now().Format("2006-01-02T15:04:05")

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		boat,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update boat: %w", err)
	}
	updatedBoat, err := c.RetrieveBoatByID(ctx, result.BoatID, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated boat: %w", err)
	}
	return updatedBoat, nil
}

// CreateBoat creates a new boat
func (c *Client) CreateBoat(ctx context.Context, boat *BoatCreate, organizationID uuid.UUID, systemID string) (*Boat, error) {
	var result BoatCreateUpdateResponse
	endpoint := "/DockMaster/Boats/UpdateBoat"

	// Initialize empty arrays if nil to satisfy DME validation
	if boat.Motors == nil {
		boat.Motors = []Motor{}
	}
	if boat.Drives == nil {
		boat.Drives = []Drive{}
	}
	if boat.Generators == nil {
		boat.Generators = []Generator{}
	}

	if boat.BoatDescriptionCodes == nil {
		boat.BoatDescriptionCodes = []BoatDescriptionCode{}
	}
	if boat.CustomInformation == nil {
		boat.CustomInformation = []CustomInformation{}
	}
	if boat.OperationsHistory == nil {
		boat.OperationsHistory = []OperationHistory{}
	}
	if boat.Attachments == nil {
		boat.Attachments = []Attachment{}
	}

	// Initialize Slip if it's nil
	if boat.Slip == (Slip{}) {
		boat.Slip = Slip{
			LastModifedDate: time.Now().Format("2006-01-02T15:04:05"),
		}
	}

	// Set LastModified if empty
	if boat.LastModified == "" {
		boat.LastModified = time.Now().Format("2006-01-02T15:04:05")
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		boat,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create boat: %w", err)
	}

	createdBoat, err := c.RetrieveBoatByID(ctx, result.BoatID, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created boat: %w", err)
	}
	return createdBoat, nil
}

// -----
// Prospect API
// -----

// SearchProspects searches for prospects
// func (c *Client) SearchProspects(ctx context.Context, searchTerm string, directHit bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
// 	var result interface{}
//
// 	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get DME credential: %w", err)
// 	}
//
// 	if credential.IsOldApi != nil && *credential.IsOldApi {
// 		payload := map[string]interface{}{
// 			"SearchString": searchTerm,
// 			"DirectHit":    directHit,
// 		}
// 		endpoint := "/Prospects/Search"
// 		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID, nil)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to search prospects: %w", err)
// 		}
// 	} else {
// 		endpoint := fmt.Sprintf("/Prospects/Search?SearchString=%s&DirectHit=%t", searchTerm, directHit)
// 		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to search prospects: %w", err)
// 		}
// 	}
// 	return result, nil
// }

// RetrieveProspectInformation retrieves prospect information
// func (c *Client) RetrieveProspectInformation(ctx context.Context, prospectID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
// 	var result interface{}
// 	endpoint := fmt.Sprintf("/Prospects/Retrieve?ProspectId=%s", prospectID)
// 	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to retrieve prospect information: %w", err)
// 	}
// 	return result, nil
// }

// -----
// Service API
// -----

// RetrieveInvoices retrieves invoices by IDs
func (c *Client) RetrieveInvoices(ctx context.Context, invoiceIDs []string, organizationID uuid.UUID, systemID string) ([]InvoiceDetailed, error) {
	var result []InvoiceDetailed
	endpoint := "/AR/RetrieveInvoices"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, invoiceIDs, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoices: %w", err)
	}
	return result, nil
}

// RetrieveCustomerInvoices retrieves invoices for a customer
func (c *Client) RetrieveCustomerInvoices(ctx context.Context, customerID string, invoiceDate string, organizationID uuid.UUID, systemID string) ([]CustomerInvoiceInquiry, error) {
	var result []CustomerInvoiceInquiry
	endpoint := fmt.Sprintf("/AR/CustomerARInquiry?CustomerId=%s&InvoiceDate=%s", customerID, invoiceDate)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer invoices: %w", err)
	}
	return result, nil
}

// RetrieveWorkOrdersList retrieves a list of work orders
// func (c *Client) RetrieveWorkOrdersList(ctx context.Context, req *WorkOrdersListRequest, organizationID uuid.UUID, systemID string) (interface{}, error) {
// 	var result interface{}
// 	endpoint := "/Service/WorkOrders/RetrieveList"
// 	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to retrieve work orders list: %w", err)
// 	}
// 	return result, nil
// }

// ListWorkOrdersForCustomer lists work orders for a customer
// func (c *Client) ListWorkOrdersForCustomer(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
// 	var result interface{}
// 	endpoint := fmt.Sprintf("/Service/WorkOrders/ListForCustomer?Status=O&CustId=%s", customerID)
// 	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list work orders for customer: %w", err)
// 	}
// 	return result, nil
// }

// -----
// Work Order API
// -----

// ListWorkOrders retrieves a full list of work orders
func (c *Client) ListWorkOrders(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) (*WorkOrderList, error) {
	var result WorkOrderList

	endpoint := fmt.Sprintf("/Service/WorkOrders/ListNewOrChanged?Page=%d&PageSize=%d", page, pageSize)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get work orders list: %w", err)
	}
	return &result, nil
}

// WorkOrderRetrieve retrieves work order information
func (c *Client) WorkOrderRetrieve(ctx context.Context, workOrderID string, detail bool, organizationID uuid.UUID, systemID string) (*WorkOrder, error) {
	var result WorkOrder

	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credential: %w", err)
	}

	var endpoint string
	if credential.IsOldApi != nil && *credential.IsOldApi {
		endpoint = fmt.Sprintf("/Service/WorkOrders/Retrieve/%s", workOrderID)
	} else {
		endpoint = fmt.Sprintf("/Service/WorkOrders/Retrieve?Id=%s&Detail=%t", workOrderID, detail)
	}

	err = c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order information: %w", err)
	}

	return &result, nil
}

// WorkOrderSearch searches for work orders
func (c *Client) WorkOrderSearch(ctx context.Context, searchTerm string, directHit string, organizationID uuid.UUID, systemID string) (*[]WorkOrderSearch, error) {
	var result []WorkOrderSearch
	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credential: %w", err)
	}

	if credential.IsOldApi != nil && *credential.IsOldApi {
		payload := map[string]interface{}{
			"SearchString": searchTerm,
			"DirectHit":    directHit,
		}
		endpoint := "/Service/WorkOrders/Search"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search work orders: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Service/WorkOrders/Search?SearchString=%s&DirectHit=%s", searchTerm, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search work orders: %w", err)
		}
	}

	return &result, nil
}

// ListWorkOrdersForCustomer lists work orders for a specific customer
func (c *Client) ListWorkOrdersForCustomer(ctx context.Context, customerID string, status string, locationCodeList string, organizationID uuid.UUID, systemID string) ([]WorkOrderShort, error) {
	var result []WorkOrderShort

	// Build the query parameters
	var queryParams []string

	// Required parameter
	queryParams = append(queryParams, fmt.Sprintf("CustId=%s", customerID))

	// Optional parameters
	if status != "" {
		queryParams = append(queryParams, fmt.Sprintf("Status=%s", status))
	}

	if locationCodeList != "" {
		queryParams = append(queryParams, fmt.Sprintf("LocationCodeList=%s", locationCodeList))
	}

	// Construct the endpoint with query parameters
	endpoint := fmt.Sprintf("/Service/WorkOrders/ListForCustomer?%s", strings.Join(queryParams, "&"))

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list work orders for customer: %w", err)
	}

	return result, nil
}

// CreateWorkOrder creates a new work order
func (c *Client) CreateWorkOrder(ctx context.Context, workOrderData map[string]interface{}, organizationID uuid.UUID, systemID string) (*WorkOrder, error) {
	var result WorkOrderCreateResponse
	endpoint := "/Service/WorkOrders/Update"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		workOrderData,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create work order: %w", err)
	}

	// Extract the work order ID from the response
	workOrderID := result.WoId
	if workOrderID == "" {
		return nil, fmt.Errorf("failed to get work order ID from response")
	}

	// Retrieve the created work order
	createdWorkOrder, err := c.WorkOrderRetrieve(ctx, workOrderID, false, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created work order: %w", err)
	}

	return createdWorkOrder, nil
}

// UpdateWorkOrder updates an existing work order
func (c *Client) UpdateWorkOrder(ctx context.Context, workOrderData map[string]interface{}, organizationID uuid.UUID, systemID string) (*WorkOrder, error) {
	var result WorkOrderCreateResponse
	endpoint := "/Service/WorkOrders/Update"

	// Ensure the work order ID is present
	workOrderID, ok := workOrderData["woId"].(string)
	if !ok || workOrderID == "" {
		return nil, fmt.Errorf("work order ID must be provided when updating")
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		workOrderData,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update work order: %w", err)
	}

	// Extract the work order ID from the response
	updatedID := result.WoId
	if updatedID == "" {
		return nil, fmt.Errorf("failed to get work order ID from response")
	}

	// Retrieve the updated work order
	updatedWorkOrder, err := c.WorkOrderRetrieve(ctx, updatedID, false, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated work order: %w", err)
	}

	return updatedWorkOrder, nil
}

// RetrieveWorkOrderOperations retrieves available operations for work orders (filtered by USE.ONLINE)
func (c *Client) RetrieveWorkOrderOperations(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) (*OperationsListResponse, error) {
	var result OperationsListResponse
	endpoint := "/Service/WorkOrders/RetrieveOperations"

	payload := PaginationRequest{
		Page:     page,
		PageSize: pageSize,
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		payload,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order operations: %w", err)
	}

	return &result, nil
}

// RetrieveAllWorkOrderOperations retrieves all operation codes (not filtered by USE.ONLINE)
func (c *Client) RetrieveAllWorkOrderOperations(ctx context.Context, page int, pageSize int, opCode string, categoryCode string, desc string, organizationID uuid.UUID, systemID string) (*OperationsListResponse, error) {
	var result OperationsListResponse
	endpoint := "/Service/WorkOrders/RetrieveAllOperations"

	payload := map[string]interface{}{
		"page":     page,
		"pageSize": pageSize,
	}

	// Add optional filter parameters if provided
	if opCode != "" {
		payload["opCode"] = opCode
	}
	if categoryCode != "" {
		payload["categoryCode"] = categoryCode
	}
	if desc != "" {
		payload["desc"] = desc
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		payload,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all work order operations: %w", err)
	}

	return &result, nil
}

// SearchAllOperations searches for operation codes by search string
func (c *Client) SearchAllOperations(ctx context.Context, searchString string, directHit bool, organizationID uuid.UUID, systemID string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	endpoint := "/Service/WorkOrders/RetrieveAllOperations/Search"

	params := make(map[string]string)
	params["SearchString"] = searchString
	params["DirectHit"] = fmt.Sprintf("%t", directHit)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search operations: %w", err)
	}

	return result, nil
}

// RetrieveCompletedWorkOrders retrieves work orders completed on a specific date
func (c *Client) RetrieveCompletedWorkOrders(ctx context.Context, completeDate string, organizationID uuid.UUID, systemID string) ([]WorkOrder, error) {
	var result []WorkOrder
	endpoint := fmt.Sprintf("/Service/WorkOrders/RetrieveCompleted?CompleteDate=%s", completeDate)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve completed work orders: %w", err)
	}

	return result, nil
}

// CreateWorkOrderFromEstimate creates a new work order from an existing estimate
func (c *Client) CreateWorkOrderFromEstimate(ctx context.Context, estimateId string, withDetail bool, withUnapprovedOps bool, organizationID uuid.UUID, systemID string) (*WorkOrderCreateResponse, error) {
	var result WorkOrderCreateResponse

	// Build the endpoint with query parameters directly
	endpoint := fmt.Sprintf("/Service/WorkOrders/CreateFromEstimate?EstimateId=%s&WithDetail=%t&WithUnapprovedOps=%t",
		estimateId, withDetail, withUnapprovedOps)

	// Make the request with POST method and query params in URL
	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		nil, // No body payload
		&result,
		organizationID,
		systemID,
		nil, // No additional params needed since they're in the URL
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create work order from estimate: %w", err)
	}

	return &result, nil
}

// DeleteOperationResponse represents the response from deleting an operation
type DeleteOperationResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

// RetrieveWorkOrderParts retrieves a list of part entries for a specific Work Order and Operation Code
func (c *Client) RetrieveWorkOrderParts(ctx context.Context, workOrderID string, opcode string, organizationID uuid.UUID, systemID string) ([]WorkOrderDetailPartEntry, error) {
	var result []WorkOrderDetailPartEntry

	params := map[string]string{
		"WorkOrderId": workOrderID,
		"OpCode":      opcode, // DockMaster API requires this parameter even if empty
	}

	endpoint := "/Service/WorkOrders/RetrieveParts"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order parts: %w", err)
	}

	return result, nil
}

// DeleteWorkOrderOperation deletes an operation from a work order
func (c *Client) DeleteWorkOrderOperation(ctx context.Context, workOrderId string, operationCode string, organizationID uuid.UUID, systemID string) (*DeleteOperationResponse, error) {
	var result DeleteOperationResponse

	// Build the endpoint with query parameters directly
	endpoint := fmt.Sprintf("/Service/WorkOrders/DeleteOperation?WorkOrder=%s&Operation=%s",
		workOrderId, operationCode)

	// Make the request with POST method and query params in URL
	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		nil, // No body payload
		&result,
		organizationID,
		systemID,
		nil, // No additional params needed since they're in the URL
	)

	if err != nil {
		return nil, fmt.Errorf("failed to delete work order operation: %w", err)
	}

	return &result, nil
}

// ListWorkOrderSublets retrieves sublet purchase orders for work orders
func (c *Client) ListWorkOrderSublets(ctx context.Context, workOrderID string, opcode string, vendorID string, organizationID uuid.UUID, systemID string) ([]SubletPurchaseOrder, error) {
	var result []SubletPurchaseOrder

	params := make(map[string]string)

	if workOrderID != "" {
		params["WorkOrderId"] = workOrderID
	}

	if opcode != "" {
		params["Opcode"] = opcode
	}

	if vendorID != "" {
		params["VendorId"] = vendorID
	}

	endpoint := "/Service/WorkOrders/Sublets/List"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list work order sublets: %w", err)
	}

	return result, nil
}

// RetrieveWorkOrderGroupDescriptions retrieves a list of work order group descriptions
func (c *Client) RetrieveWorkOrderGroupDescriptions(ctx context.Context, workOrderID string, organizationID uuid.UUID, systemID string) ([]SubletPurchaseOrder, error) {
	var result []SubletPurchaseOrder

	params := make(map[string]string)
	if workOrderID != "" {
		params["WorkOrderId"] = workOrderID
	}

	endpoint := "/Service/WorkOrders/RetrieveGroupDescription"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order group descriptions: %w", err)
	}

	return result, nil
}

// SubmitWorkOrderPartEntry submits a part entry
func (c *Client) SubmitWorkOrderPartEntry(ctx context.Context, partEntry map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/WorkOrders/SubmitPartEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		partEntry,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit work order part entry: %w", err)
	}

	return &result, nil
}

// SubmitWorkOrderTimeEntry submits a labor time entry for a technician
func (c *Client) SubmitWorkOrderTimeEntry(ctx context.Context, timeEntry map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/WorkOrders/SubmitTimeEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		timeEntry,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit work order time entry: %w", err)
	}

	return &result, nil
}

// SubmitWorkOrderSubletEntry submits a sublet entry for a work order
func (c *Client) SubmitWorkOrderSubletEntry(ctx context.Context, subletEntry map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/WorkOrders/SubmitSubletEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		subletEntry,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit work order sublet entry: %w", err)
	}

	return &result, nil
}

// SubmitEstimateSubletEntry submits a sublet entry for an estimate
func (c *Client) SubmitEstimateSubletEntry(ctx context.Context, subletEntry map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/Estimates/SubmitSubletEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		subletEntry,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit estimate sublet entry: %w", err)
	}

	return &result, nil
}

// SubmitEstimatePartEntry submits a part entry for an estimate
func (c *Client) SubmitEstimatePartEntry(ctx context.Context, partEntry map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/Estimates/SubmitPartEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		partEntry,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit estimate part entry: %w", err)
	}

	return &result, nil
}

// ListWorkOrderTimeEntries lists time entries for work orders
func (c *Client) ListWorkOrderTimeEntries(ctx context.Context, startDate string, endDate string, page int, pageSize int, listName string, detail bool, organizationID uuid.UUID, systemID string) (*TimeEntryListResponse, error) {
	var result TimeEntryListResponse

	params := make(map[string]string)
	params["StartDate"] = startDate
	params["Page"] = fmt.Sprintf("%d", page)
	params["PageSize"] = fmt.Sprintf("%d", pageSize)
	params["Detail"] = fmt.Sprintf("%t", detail)

	if endDate != "" {
		params["EndDate"] = endDate
	}

	if listName != "" {
		params["ListName"] = listName
	}

	endpoint := "/Service/WorkOrders/ListTimeEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list work order time entries: %w", err)
	}

	return &result, nil
}

// ListNewOrChangedWorkOrders searches for work orders created or changed as of a date
func (c *Client) ListNewOrChangedWorkOrders(ctx context.Context, lastUpdate string, page int, pageSize int, listName string, organizationID uuid.UUID, systemID string) (*WorkOrderList, error) {
	var result WorkOrderList

	params := make(map[string]string)
	params["LastUpdate"] = lastUpdate
	params["Page"] = fmt.Sprintf("%d", page)
	params["PageSize"] = fmt.Sprintf("%d", pageSize)

	if listName != "" {
		params["ListName"] = listName
	}

	endpoint := "/Service/WorkOrders/ListNewOrChanged"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list new or changed work orders: %w", err)
	}

	return &result, nil
}

// RetrieveWorkOrdersList retrieves a list of work orders with detail or summary information
func (c *Client) RetrieveWorkOrdersList(ctx context.Context, listRequest map[string]interface{}, organizationID uuid.UUID, systemID string) (*WorkOrderList, error) {
	var result WorkOrderList
	endpoint := "/Service/WorkOrders/RetrieveList"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		listRequest,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work orders list: %w", err)
	}

	return &result, nil
}

// -----
// Estimate API
// -----

// ListEstimatesForCustomer retrieves a list of basic estimate information for a specific customer
func (c *Client) ListEstimatesForCustomer(ctx context.Context, customerID string, status string, locationCodeList string, organizationID uuid.UUID, systemID string) ([]WorkOrderShort, error) {
	var result []WorkOrderShort

	// Build the query parameters
	var queryParams []string

	// Required parameter
	queryParams = append(queryParams, fmt.Sprintf("CustId=%s", customerID))

	// Optional parameters
	if status != "" {
		queryParams = append(queryParams, fmt.Sprintf("Status=%s", status))
	}

	if locationCodeList != "" {
		queryParams = append(queryParams, fmt.Sprintf("LocationCodeList=%s", locationCodeList))
	}

	// Construct the endpoint with query parameters
	endpoint := fmt.Sprintf("/Service/Estimates/ListForCustomer?%s", strings.Join(queryParams, "&"))

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list estimates for customer: %w", err)
	}

	return result, nil
}

// EstimateRetrieve retrieves estimate information
func (c *Client) EstimateRetrieve(ctx context.Context, estimateID string, detail bool, organizationID uuid.UUID, systemID string) (*WorkOrder, error) {
	var result WorkOrder

	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credential: %w", err)
	}

	var endpoint string
	if credential.IsOldApi != nil && *credential.IsOldApi {
		endpoint = fmt.Sprintf("/Service/Estimates/Retrieve/%s", estimateID)
	} else {
		endpoint = fmt.Sprintf("/Service/Estimates/Retrieve?Id=%s&Detail=%t", estimateID, detail)
	}

	err = c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve estimate information: %w", err)
	}

	return &result, nil
}

// RetrieveEstimateParts retrieves a list of part entries for a specific Estimate and Operation Code
func (c *Client) RetrieveEstimateParts(ctx context.Context, estimateID string, opcode string, organizationID uuid.UUID, systemID string) ([]WorkOrderDetailPartEntry, error) {
	var result []WorkOrderDetailPartEntry

	params := map[string]string{
		"EstimatesId": estimateID,
		"OpCode":      opcode, // DockMaster API requires this parameter even if empty
	}

	endpoint := "/Service/Estimates/RetrieveParts"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve estimate parts: %w", err)
	}

	return result, nil
}

// ListEstimateSublets retrieves sublet purchase orders for estimates
func (c *Client) ListEstimateSublets(ctx context.Context, workOrderID string, opcode string, vendorID string, organizationID uuid.UUID, systemID string) ([]SubletPurchaseOrder, error) {
	var result []SubletPurchaseOrder

	params := make(map[string]string)

	if workOrderID != "" {
		params["WorkOrderId"] = workOrderID
	}

	if opcode != "" {
		params["Opcode"] = opcode
	}

	if vendorID != "" {
		params["VendorId"] = vendorID
	}

	endpoint := "/Service/Estimates/Sublets/List"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list estimate sublets: %w", err)
	}

	return result, nil
}

// EstimateSearch searches for estimates
func (c *Client) EstimateSearch(ctx context.Context, searchTerm string, directHit string, organizationID uuid.UUID, systemID string) (*[]WorkOrderSearch, error) {
	var result []WorkOrderSearch
	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credential: %w", err)
	}

	if credential.IsOldApi != nil && *credential.IsOldApi {
		payload := map[string]interface{}{
			"SearchString": searchTerm,
			"DirectHit":    directHit,
		}
		endpoint := "/Service/Estimates/Search"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search estimates: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Service/Estimates/Search?SearchString=%s&DirectHit=%s", searchTerm, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search estimates: %w", err)
		}
	}

	return &result, nil
}

// DeleteEstimateOperation deletes an operation from an estimate
func (c *Client) DeleteEstimateOperation(ctx context.Context, estimateId string, operationCode string, organizationID uuid.UUID, systemID string) (*DeleteOperationResponse, error) {
	var result DeleteOperationResponse

	// Build the endpoint with query parameters directly
	endpoint := fmt.Sprintf("/Service/Estimates/DeleteOperation?EstimateId=%s&Operation=%s",
		estimateId, operationCode)

	// Make the request with POST method and query params in URL
	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		nil, // No body payload
		&result,
		organizationID,
		systemID,
		nil, // No additional params needed since they're in the URL
	)

	if err != nil {
		return nil, fmt.Errorf("failed to delete estimate operation: %w", err)
	}

	return &result, nil
}

// RetrieveEstimatesList retrieves a list of estimates with detail or summary information
func (c *Client) RetrieveEstimatesList(ctx context.Context, listRequest map[string]interface{}, organizationID uuid.UUID, systemID string) (*WorkOrderList, error) {
	var result WorkOrderList
	endpoint := "/Service/Estimates/RetrieveList"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		listRequest,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve estimates list: %w", err)
	}

	return &result, nil
}

// CreateEstimate creates a new estimate
func (c *Client) CreateEstimate(ctx context.Context, estimateData map[string]interface{}, organizationID uuid.UUID, systemID string) (*WorkOrder, error) {
	var result WorkOrderCreateResponse
	endpoint := "/Service/Estimates/Update"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		estimateData,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create estimate: %w", err)
	}

	// Extract the estimate ID from the response
	estimateID := result.WoId
	if estimateID == "" {
		return nil, fmt.Errorf("failed to get estimate ID from response")
	}

	// Retrieve the created estimate
	createdEstimate, err := c.EstimateRetrieve(ctx, estimateID, false, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created estimate: %w", err)
	}

	return createdEstimate, nil
}

// UpdateEstimate updates an existing estimate
func (c *Client) UpdateEstimate(ctx context.Context, estimateData map[string]interface{}, organizationID uuid.UUID, systemID string) (*WorkOrder, error) {
	var result WorkOrderCreateResponse
	endpoint := "/Service/Estimates/Update"

	// Ensure the estimate ID is present (woId is used by DME API for estimates)
	estimateID, ok := estimateData["woId"].(string)
	if !ok || estimateID == "" {
		return nil, fmt.Errorf("estimate ID must be provided when updating")
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		estimateData,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update estimate: %w", err)
	}

	// Extract the estimate ID from the response
	updatedID := result.WoId
	if updatedID == "" {
		return nil, fmt.Errorf("failed to get estimate ID from response")
	}

	// Retrieve the updated estimate
	updatedEstimate, err := c.EstimateRetrieve(ctx, updatedID, false, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated estimate: %w", err)
	}

	return updatedEstimate, nil
}

// -----
// General Service API
// -----

// ListNewOrChangedOpCodes searches for operation codes created or changed as of a date
func (c *Client) ListNewOrChangedOpCodes(ctx context.Context, lastUpdate string, page int, pageSize int, listName string, organizationID uuid.UUID, systemID string) (*OpCodeListResponse, error) {
	var result OpCodeListResponse

	params := make(map[string]string)
	params["LastUpdate"] = lastUpdate
	params["Page"] = fmt.Sprintf("%d", page)
	params["PageSize"] = fmt.Sprintf("%d", pageSize)

	if listName != "" {
		params["ListName"] = listName
	}

	endpoint := "/Service/ListNewOrChangedOpCodes"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list new or changed operation codes: %w", err)
	}

	return &result, nil
}

// ListWOCategoryCodes returns a list of work order category codes
func (c *Client) ListWOCategoryCodes(ctx context.Context, page int, pageSize int, listName string, organizationID uuid.UUID, systemID string) (*OpCodeListResponse, error) {
	var result OpCodeListResponse

	params := make(map[string]string)
	params["Page"] = fmt.Sprintf("%d", page)
	params["PageSize"] = fmt.Sprintf("%d", pageSize)

	if listName != "" {
		params["ListName"] = listName
	}

	endpoint := "/Service/ListWOCategoryCodes"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list work order category codes: %w", err)
	}

	return &result, nil
}

// ListOPCategoryCodes returns a list of operation category codes
func (c *Client) ListOPCategoryCodes(ctx context.Context, page int, pageSize int, listName string, organizationID uuid.UUID, systemID string) (*OpCodeListResponse, error) {
	var result OpCodeListResponse

	params := make(map[string]string)
	params["Page"] = fmt.Sprintf("%d", page)
	params["PageSize"] = fmt.Sprintf("%d", pageSize)

	if listName != "" {
		params["ListName"] = listName
	}

	endpoint := "/Service/ListOPCategoryCodes"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list operation category codes: %w", err)
	}

	return &result, nil
}

// RetrieveOperationDescriptions retrieves operation descriptions from the opcode template
func (c *Client) RetrieveOperationDescriptions(ctx context.Context, opcode string, organizationID uuid.UUID, systemID string) (*WorkOrderOperation, error) {
	var result WorkOrderOperation

	params := make(map[string]string)
	if opcode != "" {
		params["Opcode"] = opcode
	}

	endpoint := "/Service/OperationDesc"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve operation descriptions: %w", err)
	}

	return &result, nil
}

// ListTechnicians retrieves a list of technician records
func (c *Client) ListTechnicians(ctx context.Context, techID string, activeOnly bool, organizationID uuid.UUID, systemID string) ([]Technician, error) {
	var result []Technician

	params := make(map[string]string)

	if techID != "" {
		params["TechId"] = techID
	}

	params["ActiveOnly"] = fmt.Sprintf("%t", activeOnly)

	endpoint := "/Service/Technicians"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list technicians: %w", err)
	}

	return result, nil
}

// RetrieveSchedule retrieves schedule appointments
func (c *Client) RetrieveSchedule(ctx context.Context, locationCode string, startDate string, sessionID string, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	params := map[string]string{
		"locationCode": locationCode,
		"startDate":    startDate,
		"sessionId":    sessionID,
	}
	endpoint := "/Service/Schedule/Retrieve"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schedule: %w", err)
	}

	return &result, nil
}

// RetrieveScheduleForManager retrieves schedule appointments for a manager
func (c *Client) RetrieveScheduleForManager(ctx context.Context, locationCode string, startDate string, managerID string, sessionID string, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	params := map[string]string{
		"locationCode": locationCode,
		"startDate":    startDate,
		"managerId":    managerID,
		"sessionId":    sessionID,
	}
	endpoint := "/Service/Schedule/RetrieveForManager"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schedule for manager: %w", err)
	}

	return &result, nil
}

// RetrieveScheduleForTech retrieves schedule appointments for a technician
func (c *Client) RetrieveScheduleForTech(ctx context.Context, locationCode string, startDate string, techID string, sessionID string, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	params := map[string]string{
		"locationCode": locationCode,
		"startDate":    startDate,
		"techId":       techID,
		"sessionId":    sessionID,
	}
	endpoint := "/Service/Schedule/RetrieveForTech"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schedule for tech: %w", err)
	}

	return &result, nil
}

// RetrieveScheduleForWorkOrder retrieves schedule appointments for a work order
func (c *Client) RetrieveScheduleForWorkOrder(ctx context.Context, locationCode string, startDate string, workOrderID string, sessionID string, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	params := map[string]string{
		"locationCode": locationCode,
		"startDate":    startDate,
		"workOrderId":  workOrderID,
		"sessionId":    sessionID,
	}
	endpoint := "/Service/Schedule/RetrieveForWorkOrder"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve schedule for work order: %w", err)
	}

	return &result, nil
}

// RetrieveWorkOrderSchedule retrieves work order schedule
func (c *Client) RetrieveWorkOrderSchedule(ctx context.Context, locationCode string, workOrderID string, sessionID string, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	params := map[string]string{
		"locationCode": locationCode,
		"workOrderId":  workOrderID,
		"sessionId":    sessionID,
	}
	endpoint := "/Service/Schedule/WorkOrderSchedule"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order schedule: %w", err)
	}

	return &result, nil
}

// RetrieveOperationSchedule retrieves operation schedule
func (c *Client) RetrieveOperationSchedule(ctx context.Context, locationCode string, workOrderID string, opcode string, sessionID string, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	params := map[string]string{
		"locationCode": locationCode,
		"workOrderId":  workOrderID,
		"opcode":       opcode,
		"sessionId":    sessionID,
	}
	endpoint := "/Service/Schedule/OperationSchedule"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve operation schedule: %w", err)
	}

	return &result, nil
}

// UpdateSchedule updates schedule appointments
func (c *Client) UpdateSchedule(ctx context.Context, scheduleUpdate map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/Schedule/Update"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		scheduleUpdate,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return &result, nil
}

// ResolveMergeConflict resolves merge conflicts in schedule updates
func (c *Client) ResolveMergeConflict(ctx context.Context, conflictResolution map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/Schedule/ResolveMergeConflict"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		conflictResolution,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve merge conflict: %w", err)
	}

	return &result, nil
}

// -----
// Payment API
// -----

// InitiatePayment initiates a payment through the DMPay system
// This function creates a payment request that can be processed by DMPay
func (c *Client) InitiatePayment(ctx context.Context, customerID string, invoiceID string, amount float64, organizationID uuid.UUID, systemID string) (*PaymentInitiationResponse, error) {
	// Get location information to retrieve DMPay client ID
	locations, err := c.ListLocations(ctx, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %w", err)
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("no locations found for payment processing")
	}

	// Use the first location's DMPay client ID
	// In a real implementation, you might need logic to select the correct location
	dmPayClientID := locations[0].DMPayClientID
	if dmPayClientID == "" {
		return nil, fmt.Errorf("DMPay client ID not configured for this location")
	}

	// Create payment initiation response
	// This would typically create a payment session with DMPay
	result := &PaymentInitiationResponse{
		PaymentSessionID: generatePaymentSessionID(),
		DMPayClientID:    dmPayClientID,
		CustomerID:       customerID,
		InvoiceID:        invoiceID,
		Amount:           amount,
		Status:           "pending",
		PaymentURL:       fmt.Sprintf("https://pay.dockmaster.com/payment/%s", generatePaymentSessionID()),
	}

	return result, nil
}

// generatePaymentSessionID generates a unique payment session ID
func generatePaymentSessionID() string {
	return uuid.New().String()
}

// -----
// Inventory API
// -----

// RetrieveFuel retrieves one or all fuel inventory records
func (c *Client) RetrieveFuel(ctx context.Context, fuelID string, organizationID uuid.UUID, systemID string) ([]FuelInventory, error) {
	var result []FuelInventory

	var endpoint string
	if fuelID != "" {
		endpoint = fmt.Sprintf("/Inventory/RetrieveFuel?Id=%s", fuelID)
	} else {
		endpoint = "/Inventory/RetrieveFuel"
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve fuel inventory: %w", err)
	}

	return result, nil
}

// RetrieveOnlineBillcodeList retrieves a list of billing codes selected for use online
func (c *Client) RetrieveOnlineBillcodeList(ctx context.Context, organizationID uuid.UUID, systemID string) ([]OnlineBillcode, error) {
	var result []OnlineBillcode
	endpoint := "/Inventory/RetrieveOnlineBillcodeList"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve online billcode list: %w", err)
	}

	return result, nil
}

// RetrieveOnlinePartsList retrieves a list of inventory records selected for use online
func (c *Client) RetrieveOnlinePartsList(ctx context.Context, organizationID uuid.UUID, systemID string) ([]OnlinePart, error) {
	var result []OnlinePart
	endpoint := "/Inventory/RetrieveOnlinePartsList"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve online parts list: %w", err)
	}

	return result, nil
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// SubmitBatch submits a batch of cash receipts to DME
func (c *Client) SubmitBatch(ctx context.Context, locationCode string, cashReceipts []CashReceipt, postBatch bool, orgID uuid.UUID, systemID string) (*BatchSubmissionResponse, error) {
	// Prepare batch data
	batchData := map[string]interface{}{
		"locationCode": locationCode,
		"postBatch":    postBatch,
		"cashReceipts": cashReceipts,
	}

	// Make the API call to DME
	var dmeResponse map[string]interface{}
	endpoint := "/AR/SubmitBatch"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		batchData,
		&dmeResponse,
		orgID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit batch: %w", err)
	}

	// Convert DME response to our response format
	result := &BatchSubmissionResponse{
		BatchID:      getStringFromMap(dmeResponse, "batchId"),
		LocationCode: locationCode,
		PostBatch:    postBatch,
		PostResult:   getStringFromMap(dmeResponse, "postResult"),
		SubmittedAt:  time.Now(),
	}

	// Extract reference IDs if available
	if refIDs, ok := dmeResponse["referenceIds"].([]interface{}); ok {
		for _, refID := range refIDs {
			if refIDStr, ok := refID.(string); ok {
				result.ReferenceIDs = append(result.ReferenceIDs, refIDStr)
			}
		}
	}

	// Calculate total amount and receipt count
	totalAmount := 0.0
	for _, receipt := range cashReceipts {
		totalAmount += receipt.TotalPayment
	}
	result.TotalAmount = totalAmount
	result.ReceiptCount = len(cashReceipts)

	return result, nil
}

// RetrieveQtyInfo retrieves quantity information for a specific part at a specific location
func (c *Client) RetrieveQtyInfo(ctx context.Context, partNumber string, locationCode string, organizationID uuid.UUID, systemID string) (*PartQtyInfo, error) {
	var result PartQtyInfo
	endpoint := fmt.Sprintf("/Inventory/RetrieveQtyInfo?PartNumber=%s&LocationCode=%s", partNumber, locationCode)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve quantity info: %w", err)
	}

	return &result, nil
}

// RetrievePartsKit retrieves a parts kit record
func (c *Client) RetrievePartsKit(ctx context.Context, kitID string, organizationID uuid.UUID, systemID string) (*PartsKit, error) {
	var result PartsKit
	endpoint := fmt.Sprintf("/Inventory/PartsKits?Id=%s", kitID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve parts kit: %w", err)
	}

	return &result, nil
}

// ListPartsKits retrieves a list of all parts kits
func (c *Client) ListPartsKits(ctx context.Context, organizationID uuid.UUID, systemID string) ([]PartsKit, error) {
	var result []PartsKit
	endpoint := "/Inventory/PartsKits/List"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list parts kits: %w", err)
	}

	return result, nil
}

// RetrievePurchaseOrder retrieves a purchase order by its ID
func (c *Client) RetrievePurchaseOrder(ctx context.Context, poID string, organizationID uuid.UUID, systemID string) (*PurchaseOrder, error) {
	var result PurchaseOrder
	endpoint := fmt.Sprintf("/Inventory/PurchaseOrders?Id=%s", poID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchase order: %w", err)
	}

	return &result, nil
}

// ListPurchaseOrders retrieves a list of purchase orders by single date or date range
func (c *Client) ListPurchaseOrders(ctx context.Context, startDate string, endDate string, organizationID uuid.UUID, systemID string) ([]PurchaseOrder, error) {
	var result []PurchaseOrder

	var endpoint string
	if endDate != "" {
		endpoint = fmt.Sprintf("/Inventory/PurchaseOrders/PurchaseOrdersList?StartDate=%s&EndDate=%s", startDate, endDate)
	} else {
		endpoint = fmt.Sprintf("/Inventory/PurchaseOrders/PurchaseOrdersList?Date=%s", startDate)
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchase orders: %w", err)
	}

	return result, nil
}

// RetrieveSpecialOrder retrieves a special order
func (c *Client) RetrieveSpecialOrder(ctx context.Context, orderID string, organizationID uuid.UUID, systemID string) (*SpecialOrder, error) {
	var result SpecialOrder
	endpoint := fmt.Sprintf("/Inventory/SpecialOrders?Id=%s", orderID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve special order: %w", err)
	}

	return &result, nil
}

// ListCustomerSpecialOrders retrieves a list of special orders for a specific customer
func (c *Client) ListCustomerSpecialOrders(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) ([]SpecialOrder, error) {
	var result []SpecialOrder
	endpoint := fmt.Sprintf("/Inventory/SpecialOrders/CustomerSpecialOrders?CustomerId=%s", customerID)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list customer special orders: %w", err)
	}

	return result, nil
}

// ListReceivedSpecialOrders retrieves a list of received special orders
func (c *Client) ListReceivedSpecialOrders(ctx context.Context, organizationID uuid.UUID, systemID string) ([]SpecialOrder, error) {
	var result []SpecialOrder
	endpoint := "/Inventory/SpecialOrders/ReceivedOrders"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list received special orders: %w", err)
	}

	return result, nil
}

// SearchInventory searches the full inventory for an item
func (c *Client) SearchInventory(ctx context.Context, searchString string, directHit bool, organizationID uuid.UUID, systemID string) ([]InventorySearchResult, error) {
	var result []InventorySearchResult
	endpoint := fmt.Sprintf("/Inventory/Search?SearchString=%s&DirectHit=%t", searchString, directHit)

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search inventory: %w", err)
	}

	return result, nil
}

// FindParts attempts to find parts using various part numbers
func (c *Client) FindParts(ctx context.Context, partNumbers []string, organizationID uuid.UUID, systemID string) ([]InventoryPart, error) {
	var result []InventoryPart
	endpoint := "/Inventory/FindParts"

	// Send the array directly as per Dockmaster API specification
	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		partNumbers,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find parts: %w", err)
	}

	return result, nil
}

// RetrieveInventory retrieves inventory records from full inventory
func (c *Client) RetrieveInventory(ctx context.Context, query RetrieveInventoryQuery, organizationID uuid.UUID, systemID string) ([]InventoryPart, error) {
	var result []InventoryPart
	endpoint := "/Inventory/Retrieve"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		query,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory: %w", err)
	}

	return result, nil
}

// InvPayment represents an invoice payment within a cash receipt
type InvPayment struct {
	InvoiceID    string  `json:"invoiceId"`
	LocationCode string  `json:"locationCode"`
	DepositType  string  `json:"depositType"`
	PaymentAmt   float64 `json:"paymentAmt"`
	Description  string  `json:"description"`
	CustomerID   string  `json:"customerId"`
}

// CashReceipt represents a cash receipt for batch submission
type CashReceipt struct {
	CustomerID             string       `json:"customerId"`
	ReferenceNum           string       `json:"referenceNum"`
	PayType                string       `json:"payType"`
	TotalPayment           float64      `json:"totalPayment"`
	StatementDesc          string       `json:"statementDesc"`
	CCAuthCode             string       `json:"ccAuthCode"`
	CCTransactionID        string       `json:"ccTransactionID"`
	CCTransactionTimeStamp string       `json:"ccTransactionTimeStamp"`
	CCSurcharge            float64      `json:"ccSurcharge"`
	CCSurchargeTax         float64      `json:"ccSurchargeTax"`
	CCSurchargeTaxSchema   string       `json:"ccSurchargeTaxSchema"`
	CCSurchargeTaxIds      []string     `json:"ccSurchargeTaxIds"`
	InvPayments            []InvPayment `json:"invPayments"`
}

// BatchSubmissionResponse represents the response from DME batch submission
type BatchSubmissionResponse struct {
	BatchID      string    `json:"batchId"`
	LocationCode string    `json:"locationCode"`
	PostBatch    bool      `json:"postBatch"`
	ReferenceIDs []string  `json:"referenceIds"`
	PostResult   string    `json:"postResult"`
	SubmittedAt  time.Time `json:"submittedAt"`
	TotalAmount  float64   `json:"totalAmount"`
	ReceiptCount int       `json:"receiptCount"`
}

// -----
// Vendor API
// -----

// ListVendors retrieves a list of all vendors
func (c *Client) ListVendors(ctx context.Context, organizationID uuid.UUID, systemID string) ([]Vendor, error) {
	var result []Vendor
	endpoint := "/Vendors/List"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list vendors: %w", err)
	}

	return result, nil
}

// SearchVendors searches for vendors based on search string
func (c *Client) SearchVendors(ctx context.Context, searchString string, directHit bool, organizationID uuid.UUID, systemID string) ([]VendorSearchResult, error) {
	var result []VendorSearchResult
	endpoint := "/Vendors/Search"

	params := map[string]string{
		"SearchString": searchString,
		"DirectHit":    fmt.Sprintf("%t", directHit),
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search vendors: %w", err)
	}

	return result, nil
}

// RetrieveVendor retrieves a single vendor by ID
func (c *Client) RetrieveVendor(ctx context.Context, vendorID string, organizationID uuid.UUID, systemID string) (*Vendor, error) {
	var result Vendor
	endpoint := "/Vendors/RetrieveVendor"

	params := map[string]string{
		"VendorId": vendorID,
	}

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve vendor: %w", err)
	}

	return &result, nil
}

// NextReferenceNumber retrieves the next AR reference number for a customer
func (c *Client) NextReferenceNumber(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (string, error) {
	// The DME endpoint returns the next available AR reference number for a given customer
	endpoint := fmt.Sprintf("/AR/NextReferenceNumber?CustomerId=%s", customerID)

	// Decode into a generic map to be resilient to response shape differences
	var result map[string]interface{}
	if err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil); err != nil {
		return "", fmt.Errorf("failed to retrieve next reference number: %w", err)
	}

	// Try common keys
	keys := []string{"referenceNumber", "ReferenceNumber", "reference", "Reference"}
	for _, k := range keys {
		if v, ok := result[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s, nil
			}
		}
	}

	// If the API returned a bare string, attempt to handle that as well
	if v, ok := result["value"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, nil
		}
	}

	return "", fmt.Errorf("reference number not found in response")
}

// RetrievePayTypes retrieves available pay types from DME
func (c *Client) RetrievePayTypes(ctx context.Context, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/General/PayTypes/RetrievePayTypes"

	if err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to retrieve pay types: %w", err)
	}

	return &result, nil
}

// SubmitEstimateLaborEntry submits a labor entry for an estimate
func (c *Client) SubmitEstimateLaborEntry(ctx context.Context, laborEntry map[string]interface{}, organizationID uuid.UUID, systemID string) (*interface{}, error) {
	var result interface{}
	endpoint := "/Service/Estimates/SubmitLaborEntry"

	err := c.DoJSONRequest(
		ctx,
		http.MethodPost,
		endpoint,
		laborEntry,
		&result,
		organizationID,
		systemID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit estimate labor entry: %w", err)
	}

	return &result, nil
}

// RetrieveEstimateLabor retrieves labor entries for an estimate
func (c *Client) RetrieveEstimateLabor(ctx context.Context, estimateID string, opcode string, organizationID uuid.UUID, systemID string) ([]LaborEntry, error) {
	var result []LaborEntry

	params := map[string]string{
		"EstimatesId": estimateID,
		"OpCode":      opcode, // DockMaster API requires this parameter even if empty
	}

	endpoint := "/Service/Estimates/RetrieveLabor"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve estimate labor: %w", err)
	}

	return result, nil
}

// RetrieveWorkOrderLaborDetail retrieves labor detail for a work order
func (c *Client) RetrieveWorkOrderLaborDetail(ctx context.Context, workOrderID string, opcode string, organizationID uuid.UUID, systemID string) ([]LaborEntry, error) {
	var result []LaborEntry

	params := map[string]string{
		"WorkOrderId": workOrderID,
		"OpCode":      opcode, // DockMaster API requires this parameter even if empty
	}

	endpoint := "/Service/WorkOrderLaborDetail"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order labor detail: %w", err)
	}

	return result, nil
}

// RetrieveWorkOrderPartDetail retrieves detailed part information for a work order operation
func (c *Client) RetrieveWorkOrderPartDetail(ctx context.Context, workOrderID string, opcode string, organizationID uuid.UUID, systemID string) ([]WorkOrderPartDetail, error) {
	var result []WorkOrderPartDetail

	params := map[string]string{
		"WodID":  workOrderID,
		"OpCode": opcode,
	}

	endpoint := "/Service/WorkOrderPartDetail"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order part detail: %w", err)
	}

	return result, nil
}

// RetrieveWorkOrderLaborDetailRecords retrieves comprehensive individual labor detail records for a work order operation
func (c *Client) RetrieveWorkOrderLaborDetailRecords(ctx context.Context, workOrderID string, opcode string, organizationID uuid.UUID, systemID string) ([]WorkOrderLaborDetailRecord, error) {
	var result []WorkOrderLaborDetailRecord

	params := map[string]string{
		"WorkOrderId": workOrderID,
		"OpCode":      opcode,
	}

	endpoint := "/Service/WorkOrderLaborDetail"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work order labor detail records: %w", err)
	}

	return result, nil
}

// RetrieveEstimatePartDetail retrieves detailed part information for an estimate operation
func (c *Client) RetrieveEstimatePartDetail(ctx context.Context, estimateID string, opcode string, organizationID uuid.UUID, systemID string) ([]WorkOrderPartDetail, error) {
	var result []WorkOrderPartDetail

	params := map[string]string{
		"WodID":    estimateID,
		"OpCode":   opcode,
		"Estimate": "true",
	}

	endpoint := "/Service/WorkOrderPartDetail"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve estimate part detail: %w", err)
	}

	return result, nil
}

// RetrieveEstimateLaborDetailRecords retrieves comprehensive individual labor detail records for an estimate operation
func (c *Client) RetrieveEstimateLaborDetailRecords(ctx context.Context, estimateID string, opcode string, organizationID uuid.UUID, systemID string) ([]WorkOrderLaborDetailRecord, error) {
	var result []WorkOrderLaborDetailRecord

	params := map[string]string{
		"WorkOrderId": estimateID,
		"OpCode":      opcode,
	}

	endpoint := "/Service/WorkOrderLaborDetail"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve estimate labor detail records: %w", err)
	}

	return result, nil
}
