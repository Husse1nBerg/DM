package dme

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// -----
// General API
// -----

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
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
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
func (c *Client) CustomersList(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) (*CustomerList, error) {
	var result CustomerList

	endpoint := fmt.Sprintf("/Customers/ListNewOrChanged?Page=%d&PageSize=%d", page, pageSize)
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
func (c *Client) BoatsList(ctx context.Context, page int, pageSize int, organizationID uuid.UUID, systemID string) ([]Boat, error) {
	var result BoatList
	endpoint := fmt.Sprintf("/Boats/ListNewOrChanged?Page=%d&PageSize=%d", page, pageSize)

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

	return result.Content, nil
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

// UpdateBoat updates a boat
func (c *Client) UpdateBoat(ctx context.Context, boat *BoatUpdate, organizationID uuid.UUID, systemID string) (*Boat, error) {
	var result Boat
	endpoint := "/DockMaster/Boats/UpdateBoat"

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

	return &result, nil
}

// CreateBoat creates a new boat
func (c *Client) CreateBoat(ctx context.Context, boat *BoatCreate, organizationID uuid.UUID, systemID string) (*Boat, error) {
	var result BoatCreateUpdateResponse
	endpoint := "/DockMaster/Boats/UpdateBoat"

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
// func (c *Client) RetrieveInvoices(ctx context.Context, invoiceIDs []string, organizationID uuid.UUID, systemID string) (interface{}, error) {
// 	var result interface{}
// 	endpoint := "/AR/RetrieveInvoices"
// 	err := c.DoJSONRequest(ctx, http.MethodPost, endpoint, invoiceIDs, &result, organizationID, systemID, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to retrieve invoices: %w", err)
// 	}
// 	return result, nil
// }

// RetrieveCustomerInvoices retrieves invoices for a customer
// func (c *Client) RetrieveCustomerInvoices(ctx context.Context, customerID string, invoiceDate string, organizationID uuid.UUID, systemID string) (interface{}, error) {
// 	var result interface{}
// 	endpoint := fmt.Sprintf("/AR/CustomerARInquiry?CustomerId=%s&InvoiceDate=%s", customerID, invoiceDate)
// 	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result, organizationID, systemID, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to retrieve customer invoices: %w", err)
// 	}
// 	return result, nil
// }

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
