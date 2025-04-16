package dme

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// GetVersionInformation retrieves version information
func (c *Client) GetVersionInformation(ctx context.Context, organizationID uuid.UUID, systemID string) (*VersionInfo, error) {
	var result VersionInfo
	endpoint := fmt.Sprintf("/General/VersionInformation")
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version info: %w", err)
	}
	return &result, nil
}

// ReversePayment reverses a payment
func (c *Client) ReversePayment(ctx context.Context, req *ReversePaymentRequest, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/AR/ReversePayment"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse payment: %w", err)
	}
	return result, nil
}

// RetrieveCustomerContracts retrieves a customer's contracts
func (c *Client) RetrieveCustomerContracts(ctx context.Context, id string, isProspect bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	idType := "ProspectId"
	if !isProspect {
		idType = "CustomerId"
	}
	endpoint := fmt.Sprintf("/UnitSales/RetrieveCustomerContracts?%s=%s", idType, id)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer contracts: %w", err)
	}
	return result, nil
}

// RetrieveCustomerQuotes retrieves a customer's quotes
func (c *Client) RetrieveCustomerQuotes(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/UnitSales/RetrieveCustomerQuotes?CustomerId=%s", customerID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer quotes: %w", err)
	}
	return result, nil
}

// GetCustomerSpecialOrders retrieves special orders for a customer
func (c *Client) GetCustomerSpecialOrders(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/Inventory/SpecialOrders/CustomerSpecialOrders?CustomerId=%s&IncludeClosed=false", customerID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer special orders: %w", err)
	}
	return result, nil
}

// RetrieveWorkOrdersList retrieves a list of work orders
func (c *Client) RetrieveWorkOrdersList(ctx context.Context, req *WorkOrdersListRequest, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/Service/WorkOrders/RetrieveList"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve work orders list: %w", err)
	}
	return result, nil
}

// CreateMiscARCharge creates a miscellaneous AR charge
func (c *Client) CreateMiscARCharge(ctx context.Context, req *MiscARCharge, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/AR/CreateMiscARCharge"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to create misc AR charge: %w", err)
	}
	return result, nil
}

// SearchCustomers searches for customers
func (c *Client) SearchCustomers(ctx context.Context, searchTerm string, directHit bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	if c.config.IsOldAPI {
		payload := map[string]interface{}{
			"SearchString": searchTerm,
			"DirectHit":    directHit,
		}
		endpoint := "/Customers/Search"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to search customers: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Customers/Search?SearchString=%s&DirectHit=%t", searchTerm, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to search customers: %w", err)
		}
	}

	return result, nil
}

// GetShortCustomersList retrieves a short list of customers
func (c *Client) GetShortCustomersList(ctx context.Context, searchString string, directHit bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	if c.config.IsOldAPI {
		payload := map[string]interface{}{
			"SearchString": searchString,
			"DirectHit":    directHit,
		}
		endpoint := "/Customers/List"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get short customers list: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Customers/List?SearchString=%s&DirectHit=%t", searchString, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get short customers list: %w", err)
		}
	}

	return result, nil
}

// GetFullCustomersList retrieves a full list of customers
func (c *Client) GetFullCustomersList(ctx context.Context, searchString string, directHit bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	if c.config.IsOldAPI {
		payload := map[string]interface{}{
			"SearchString": searchString,
			"DirectHit":    directHit,
		}
		endpoint := "/Customers/RetrieveCustomers"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get full customers list: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Customers/RetrieveCustomers?SearchString=%s&DirectHit=%t", searchString, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get full customers list: %w", err)
		}
	}

	return result, nil
}

// SearchProspects searches for prospects
func (c *Client) SearchProspects(ctx context.Context, searchTerm string, directHit bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	if c.config.IsOldAPI {
		payload := map[string]interface{}{
			"SearchString": searchTerm,
			"DirectHit":    directHit,
		}
		endpoint := "/Prospects/Search"
		err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, payload, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to search prospects: %w", err)
		}
	} else {
		endpoint := fmt.Sprintf("/Prospects/Search?SearchString=%s&DirectHit=%t", searchTerm, directHit)
		err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
		if err != nil {
			return nil, fmt.Errorf("failed to search prospects: %w", err)
		}
	}

	return result, nil
}

// RetrieveCustomerInformation retrieves customer information
func (c *Client) RetrieveCustomerInformation(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	var endpoint string
	if c.config.IsOldAPI {
		endpoint = fmt.Sprintf("/Customers/RetrieveCustomer/%s", customerID)
	} else {
		endpoint = fmt.Sprintf("/Customers/RetrieveCustomer?CustomerId=%s", customerID)
	}

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer information: %w", err)
	}

	return result, nil
}

// RetrieveProspectInformation retrieves prospect information
func (c *Client) RetrieveProspectInformation(ctx context.Context, prospectID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/Prospects/Retrieve?ProspectId=%s", prospectID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve prospect information: %w", err)
	}
	return result, nil
}

// ListWorkOrdersForCustomer lists work orders for a customer
func (c *Client) ListWorkOrdersForCustomer(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/Service/WorkOrders/ListForCustomer?Status=O&CustId=%s", customerID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list work orders for customer: %w", err)
	}
	return result, nil
}

// RetrieveBoatsForCustomer retrieves boats associated with a customer
func (c *Client) RetrieveBoatsForCustomer(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/Boats/RetrieveBoats?CustId=%s", customerID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve boats for customer: %w", err)
	}
	return result, nil
}

// RetrieveBoatByID retrieves a boat by its ID
func (c *Client) RetrieveBoatByID(ctx context.Context, boatID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/Boats/RetrieveBoat?BoatId=%s", boatID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve boat by ID: %w", err)
	}
	return result, nil
}

// RetrieveInvoices retrieves invoices by IDs
func (c *Client) RetrieveInvoices(ctx context.Context, invoiceIDs []string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/AR/RetrieveInvoices"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, invoiceIDs, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoices: %w", err)
	}
	return result, nil
}

// ApplyCreditMemo applies a credit memo
func (c *Client) ApplyCreditMemo(ctx context.Context, req *CreditMemoApplication, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/AR/ApplyCreditMemo"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to apply credit memo: %w", err)
	}
	return result, nil
}

// SubmitBatch submits a batch
func (c *Client) SubmitBatch(ctx context.Context, req *BatchSubmissionRequest, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/AR/SubmitBatch"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to submit batch: %w", err)
	}
	return result, nil
}

// RetrieveClerkByID retrieves a clerk by ID
func (c *Client) RetrieveClerkByID(ctx context.Context, clerkID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	var endpoint string
	if c.config.IsOldAPI {
		endpoint = fmt.Sprintf("/api/v2/General/Clerks/%s", clerkID)
	} else {
		endpoint = fmt.Sprintf("/General/Clerks?ClerkId=%s", clerkID)
	}

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve clerk by ID: %w", err)
	}
	return result, nil
}

// ListClerks lists all clerks
func (c *Client) ListClerks(ctx context.Context, includeInactive bool, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/General/Clerks/List?IncludeInactive=%t", includeInactive)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list clerks: %w", err)
	}
	return result, nil
}

// RetrieveContract retrieves a contract by ID
func (c *Client) RetrieveContract(ctx context.Context, contractID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	var endpoint string
	if c.config.IsOldAPI {
		endpoint = fmt.Sprintf("/UnitSales/RetrieveContract/%s", contractID)
	} else {
		endpoint = fmt.Sprintf("/UnitSales/RetrieveContract?ContractId=%s", contractID)
	}

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve contract: %w", err)
	}
	return result, nil
}

// RetrieveReservation retrieves a reservation by ID
func (c *Client) RetrieveReservation(ctx context.Context, reservationID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}

	var endpoint string
	if c.config.IsOldAPI {
		endpoint = fmt.Sprintf("/MarinaOps/Reservations/%s", reservationID)
	} else {
		endpoint = fmt.Sprintf("/MarinaOps/Reservations?ReservationId=%s", reservationID)
	}

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve reservation: %w", err)
	}
	return result, nil
}

// RetrievePayTypes retrieves payment types
func (c *Client) RetrievePayTypes(ctx context.Context, organizationID uuid.UUID, systemID string) ([]PayType, error) {
	var result []PayType
	endpoint := "/General/PayTypes/RetrievePayTypes"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pay types: %w", err)
	}
	return result, nil
}

// ListLocations lists all locations
func (c *Client) ListLocations(ctx context.Context, organizationID uuid.UUID, systemID string) ([]Location, error) {
	var result []Location
	endpoint := "/General/Locations/"

	err := c.DoJSONRequest(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
		&result,
		organizationID,
		systemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	return result, nil
}

// RetrieveCustomerInvoices retrieves invoices for a customer
func (c *Client) RetrieveCustomerInvoices(ctx context.Context, customerID string, invoiceDate string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := fmt.Sprintf("/AR/CustomerARInquiry?CustomerId=%s&InvoiceDate=%s", customerID, invoiceDate)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer invoices: %w", err)
	}
	return result, nil
}

// GetNextReferenceNumber gets the next reference number for a customer
func (c *Client) GetNextReferenceNumber(ctx context.Context, customerID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	if c.config.IsOldAPI {
		return nil, fmt.Errorf("operation not supported with old API version")
	}

	var result interface{}
	endpoint := fmt.Sprintf("/AR/NextReferenceNumber?CustomerId=%s", customerID)
	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next reference number: %w", err)
	}
	return result, nil
}

// CalculateTax calculates tax for a sale
func (c *Client) CalculateTax(ctx context.Context, req *TaxCalculationRequest, organizationID uuid.UUID, systemID string) (interface{}, error) {
	var result interface{}
	endpoint := "/AR/CalculateTax"
	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, req, &result, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tax: %w", err)
	}
	return result, nil
}

// CalculateTaxWithPayType calculates tax for a sale based on pay type
func (c *Client) CalculateTaxWithPayType(ctx context.Context, referenceNumber, customerID, locationCode string, saleAmount float64, taxDate, payTypeID string, organizationID uuid.UUID, systemID string) (interface{}, error) {
	// First get the pay types
	payTypes, err := c.RetrievePayTypes(ctx, organizationID, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pay types: %w", err)
	}

	// Find the tax schema for the pay type
	var taxSchema string
	for _, payType := range payTypes {
		if payType.ID == payTypeID {
			taxSchema = payType.CCSurchargeTaxSchema
			break
		}
	}

	if taxSchema == "" {
		return nil, fmt.Errorf("pay type not found or no tax schema available")
	}

	// Calculate tax with the found schema
	taxReq := &TaxCalculationRequest{
		ReferenceNumber: referenceNumber,
		CustomerId:      customerID,
		LocationCode:    locationCode,
		SaleAmount:      saleAmount,
		TaxDate:         taxDate,
		TaxSchema:       taxSchema,
	}

	return c.CalculateTax(ctx, taxReq, organizationID, systemID)
}

// VesselDetails represents vessel information from DME
type VesselDetails struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Length       int    `json:"length"`
	Beam         int    `json:"beam"`
	Registration string `json:"registration"`
	Type         string `json:"type"`
	OwnerID      string `json:"ownerId"`
}

// GetVessels retrieves vessels from the DME API
func (c *Client) GetVessels(ctx context.Context, organizationID uuid.UUID, systemID string) ([]VesselDetails, error) {
	var response struct {
		Vessels []VesselDetails `json:"vessels"`
	}

	err := c.DoJSONRequest(
		ctx,
		"GET",
		"/vessels",
		nil,
		&response,
		organizationID,
		systemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get vessels: %w", err)
	}

	return response.Vessels, nil
}

// GetVesselByID retrieves a vessel by ID from the DME API
func (c *Client) GetVesselByID(ctx context.Context, vesselID string, organizationID uuid.UUID, systemID string) (*VesselDetails, error) {
	var vessel VesselDetails

	err := c.DoJSONRequest(
		ctx,
		"GET",
		"/vessels/"+vesselID,
		nil,
		&vessel,
		organizationID,
		systemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get vessel %s: %w", vesselID, err)
	}

	return &vessel, nil
}
