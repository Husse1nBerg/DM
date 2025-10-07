package responses

import "github.com/dockworks/dm-web-backend/pkg/dme"

// FuelInventoryResponse represents the response for fuel inventory
type FuelInventoryResponse struct {
	ID               string  `json:"id"`
	Description      string  `json:"description"`
	LocationCode     string  `json:"locationCode"`
	TankNumber       string  `json:"tankNumber"`
	TankCapacity     float64 `json:"tankCapacity"`
	CurrentQuantity  float64 `json:"currentQuantity"`
	UnitOfMeasure    string  `json:"unitOfMeasure"`
	CostPerUnit      float64 `json:"costPerUnit"`
	PricePerUnit     float64 `json:"pricePerUnit"`
	ReorderLevel     float64 `json:"reorderLevel"`
	LastDeliveryDate string  `json:"lastDeliveryDate"`
	LastDeliveryQty  float64 `json:"lastDeliveryQty"`
	LastModified     string  `json:"lastModified"`
}

// NewFuelInventoryResponse creates a response from DME FuelInventory model
func NewFuelInventoryResponse(fuel dme.FuelInventory) FuelInventoryResponse {
	return FuelInventoryResponse{
		ID:               fuel.ID,
		Description:      fuel.Description,
		LocationCode:     fuel.LocationCode,
		TankNumber:       fuel.TankNumber,
		TankCapacity:     fuel.TankCapacity,
		CurrentQuantity:  fuel.CurrentQuantity,
		UnitOfMeasure:    fuel.UnitOfMeasure,
		CostPerUnit:      fuel.CostPerUnit,
		PricePerUnit:     fuel.PricePerUnit,
		ReorderLevel:     fuel.ReorderLevel,
		LastDeliveryDate: fuel.LastDeliveryDate,
		LastDeliveryQty:  fuel.LastDeliveryQty,
		LastModified:     fuel.LastModified,
	}
}

// OnlineBillcodeResponse represents the response for an online billcode
type OnlineBillcodeResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Department  string  `json:"department"`
	Price       float64 `json:"price"`
	TaxFlag     bool    `json:"taxFlag"`
	Active      bool    `json:"active"`
}

// NewOnlineBillcodeResponse creates a response from DME OnlineBillcode model
func NewOnlineBillcodeResponse(billcode dme.OnlineBillcode) OnlineBillcodeResponse {
	return OnlineBillcodeResponse{
		ID:          billcode.ID,
		Description: billcode.Description,
		Department:  billcode.Department,
		Price:       billcode.Price,
		TaxFlag:     billcode.TaxFlag,
		Active:      billcode.Active,
	}
}

// OnlinePartResponse represents the response for an online part
type OnlinePartResponse struct {
	ID                string  `json:"id"`
	PartNumber        string  `json:"partNumber"`
	Description       string  `json:"description"`
	LocationCode      string  `json:"locationCode"`
	QuantityOnHand    float64 `json:"quantityOnHand"`
	QuantityAvailable float64 `json:"quantityAvailable"`
	Price             float64 `json:"price"`
	Cost              float64 `json:"cost"`
	UnitOfMeasure     string  `json:"unitOfMeasure"`
	VendorID          string  `json:"vendorId"`
	VendorName        string  `json:"vendorName"`
	TaxFlag           bool    `json:"taxFlag"`
	Active            bool    `json:"active"`
	Department        string  `json:"department"`
	LastModified      string  `json:"lastModified"`
}

// NewOnlinePartResponse creates a response from DME OnlinePart model
func NewOnlinePartResponse(part dme.OnlinePart) OnlinePartResponse {
	return OnlinePartResponse{
		ID:                part.ID,
		PartNumber:        part.PartNumber,
		Description:       part.Description,
		LocationCode:      part.LocationCode,
		QuantityOnHand:    part.QuantityOnHand,
		QuantityAvailable: part.QuantityAvailable,
		Price:             part.Price,
		Cost:              part.Cost,
		UnitOfMeasure:     part.UnitOfMeasure,
		VendorID:          part.VendorID,
		VendorName:        part.VendorName,
		TaxFlag:           part.TaxFlag,
		Active:            part.Active,
		Department:        part.Department,
		LastModified:      part.LastModified,
	}
}

// PartQtyInfoResponse represents the response for part quantity information
type PartQtyInfoResponse struct {
	PartNumber        string  `json:"partNumber"`
	LocationCode      string  `json:"locationCode"`
	QuantityOnHand    float64 `json:"quantityOnHand"`
	QuantityOnOrder   float64 `json:"quantityOnOrder"`
	QuantityCommitted float64 `json:"quantityCommitted"`
	QuantityAvailable float64 `json:"quantityAvailable"`
	Cost              float64 `json:"cost"`
	AverageCost       float64 `json:"averageCost"`
	LastCost          float64 `json:"lastCost"`
	ReorderLevel      float64 `json:"reorderLevel"`
	ReorderQty        float64 `json:"reorderQty"`
	MinOrderQty       float64 `json:"minOrderQty"`
	MaxOrderQty       float64 `json:"maxOrderQty"`
	LastModified      string  `json:"lastModified"`
}

// NewPartQtyInfoResponse creates a response from DME PartQtyInfo model
func NewPartQtyInfoResponse(info dme.PartQtyInfo) PartQtyInfoResponse {
	return PartQtyInfoResponse{
		PartNumber:        info.PartNumber,
		LocationCode:      info.LocationCode,
		QuantityOnHand:    info.QuantityOnHand,
		QuantityOnOrder:   info.QuantityOnOrder,
		QuantityCommitted: info.QuantityCommitted,
		QuantityAvailable: info.QuantityAvailable,
		Cost:              info.Cost,
		AverageCost:       info.AverageCost,
		LastCost:          info.LastCost,
		ReorderLevel:      info.ReorderLevel,
		ReorderQty:        info.ReorderQty,
		MinOrderQty:       info.MinOrderQty,
		MaxOrderQty:       info.MaxOrderQty,
		LastModified:      info.LastModified,
	}
}

// PartsKitItemResponse represents the response for a parts kit item
type PartsKitItemResponse struct {
	PartNumber  string  `json:"partNumber"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`
	Cost        float64 `json:"cost"`
}

// PartsKitResponse represents the response for a parts kit
type PartsKitResponse struct {
	ID           string                 `json:"id"`
	Description  string                 `json:"description"`
	Active       bool                   `json:"active"`
	TotalPrice   float64                `json:"totalPrice"`
	TotalCost    float64                `json:"totalCost"`
	Items        []PartsKitItemResponse `json:"items"`
	LastModified string                 `json:"lastModified"`
}

// NewPartsKitResponse creates a response from DME PartsKit model
func NewPartsKitResponse(kit dme.PartsKit) PartsKitResponse {
	items := make([]PartsKitItemResponse, len(kit.Items))
	for i, item := range kit.Items {
		items[i] = PartsKitItemResponse{
			PartNumber:  item.PartNumber,
			Description: item.Description,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Cost:        item.Cost,
		}
	}

	return PartsKitResponse{
		ID:           kit.ID,
		Description:  kit.Description,
		Active:       kit.Active,
		TotalPrice:   kit.TotalPrice,
		TotalCost:    kit.TotalCost,
		Items:        items,
		LastModified: kit.LastModified,
	}
}

// PurchaseOrderLineResponse represents the response for a purchase order line
type PurchaseOrderLineResponse struct {
	LineNumber    int     `json:"lineNumber"`
	PartNumber    string  `json:"partNumber"`
	Description   string  `json:"description"`
	QuantityOrder float64 `json:"quantityOrder"`
	QuantityRecvd float64 `json:"quantityRecvd"`
	UnitCost      float64 `json:"unitCost"`
	ExtendedCost  float64 `json:"extendedCost"`
	TaxFlag       bool    `json:"taxFlag"`
	Discount      float64 `json:"discount"`
}

// PurchaseOrderResponse represents the response for a purchase order
type PurchaseOrderResponse struct {
	ID              string                      `json:"id"`
	PONumber        string                      `json:"poNumber"`
	VendorID        string                      `json:"vendorId"`
	VendorName      string                      `json:"vendorName"`
	OrderDate       string                      `json:"orderDate"`
	ExpectedDate    string                      `json:"expectedDate"`
	ReceivedDate    string                      `json:"receivedDate"`
	Status          string                      `json:"status"`
	LocationCode    string                      `json:"locationCode"`
	TotalAmount     float64                     `json:"totalAmount"`
	TotalReceived   float64                     `json:"totalReceived"`
	Comments        string                      `json:"comments"`
	Lines           []PurchaseOrderLineResponse `json:"lines"`
	CreatedBy       string                      `json:"createdBy"`
	LastModifiedBy  string                      `json:"lastModifiedBy"`
	LastModified    string                      `json:"lastModified"`
}

// NewPurchaseOrderResponse creates a response from DME PurchaseOrder model
func NewPurchaseOrderResponse(po dme.PurchaseOrder) PurchaseOrderResponse {
	lines := make([]PurchaseOrderLineResponse, len(po.Lines))
	for i, line := range po.Lines {
		lines[i] = PurchaseOrderLineResponse{
			LineNumber:    line.LineNumber,
			PartNumber:    line.PartNumber,
			Description:   line.Description,
			QuantityOrder: line.QuantityOrder,
			QuantityRecvd: line.QuantityRecvd,
			UnitCost:      line.UnitCost,
			ExtendedCost:  line.ExtendedCost,
			TaxFlag:       line.TaxFlag,
			Discount:      line.Discount,
		}
	}

	return PurchaseOrderResponse{
		ID:             po.ID,
		PONumber:       po.PONumber,
		VendorID:       po.VendorID,
		VendorName:     po.VendorName,
		OrderDate:      po.OrderDate,
		ExpectedDate:   po.ExpectedDate,
		ReceivedDate:   po.ReceivedDate,
		Status:         po.Status,
		LocationCode:   po.LocationCode,
		TotalAmount:    po.TotalAmount,
		TotalReceived:  po.TotalReceived,
		Comments:       po.Comments,
		Lines:          lines,
		CreatedBy:      po.CreatedBy,
		LastModifiedBy: po.LastModifiedBy,
		LastModified:   po.LastModified,
	}
}

// SpecialOrderResponse represents the response for a special order
type SpecialOrderResponse struct {
	ID              string  `json:"id"`
	OrderNumber     string  `json:"orderNumber"`
	CustomerID      string  `json:"customerId"`
	CustomerName    string  `json:"customerName"`
	PartNumber      string  `json:"partNumber"`
	Description     string  `json:"description"`
	QuantityOrdered float64 `json:"quantityOrdered"`
	QuantityRecvd   float64 `json:"quantityRecvd"`
	UnitPrice       float64 `json:"unitPrice"`
	UnitCost        float64 `json:"unitCost"`
	OrderDate       string  `json:"orderDate"`
	ExpectedDate    string  `json:"expectedDate"`
	ReceivedDate    string  `json:"receivedDate"`
	Status          string  `json:"status"`
	VendorID        string  `json:"vendorId"`
	VendorName      string  `json:"vendorName"`
	LocationCode    string  `json:"locationCode"`
	Comments        string  `json:"comments"`
	PONumber        string  `json:"poNumber"`
	LastModified    string  `json:"lastModified"`
}

// NewSpecialOrderResponse creates a response from DME SpecialOrder model
func NewSpecialOrderResponse(order dme.SpecialOrder) SpecialOrderResponse {
	return SpecialOrderResponse{
		ID:              order.ID,
		OrderNumber:     order.OrderNumber,
		CustomerID:      order.CustomerID,
		CustomerName:    order.CustomerName,
		PartNumber:      order.PartNumber,
		Description:     order.Description,
		QuantityOrdered: order.QuantityOrdered,
		QuantityRecvd:   order.QuantityRecvd,
		UnitPrice:       order.UnitPrice,
		UnitCost:        order.UnitCost,
		OrderDate:       order.OrderDate,
		ExpectedDate:    order.ExpectedDate,
		ReceivedDate:    order.ReceivedDate,
		Status:          order.Status,
		VendorID:        order.VendorID,
		VendorName:      order.VendorName,
		LocationCode:    order.LocationCode,
		Comments:        order.Comments,
		PONumber:        order.PONumber,
		LastModified:    order.LastModified,
	}
}

// InventorySearchResultResponse represents the response for inventory search result
type InventorySearchResultResponse struct {
	PartNumber        string  `json:"partNumber"`
	Description       string  `json:"description"`
	LocationCode      string  `json:"locationCode"`
	QuantityOnHand    float64 `json:"quantityOnHand"`
	QuantityAvailable float64 `json:"quantityAvailable"`
	Price             float64 `json:"price"`
	Cost              float64 `json:"cost"`
	UnitOfMeasure     string  `json:"unitOfMeasure"`
	Department        string  `json:"department"`
	VendorID          string  `json:"vendorId"`
	VendorName        string  `json:"vendorName"`
}

// NewInventorySearchResultResponse creates a response from DME InventorySearchResult model
func NewInventorySearchResultResponse(result dme.InventorySearchResult) InventorySearchResultResponse {
	return InventorySearchResultResponse{
		PartNumber:        result.PartNumber,
		Description:       result.Description,
		LocationCode:      result.LocationCode,
		QuantityOnHand:    result.QuantityOnHand,
		QuantityAvailable: result.QuantityAvailable,
		Price:             result.Price,
		Cost:              result.Cost,
		UnitOfMeasure:     result.UnitOfMeasure,
		Department:        result.Department,
		VendorID:          result.VendorID,
		VendorName:        result.VendorName,
	}
}

// InventoryPartResponse represents the response for a full inventory part
type InventoryPartResponse struct {
	PartNumber           string  `json:"partNumber"`
	Description          string  `json:"description"`
	LocationCode         string  `json:"locationCode"`
	QuantityOnHand       float64 `json:"quantityOnHand"`
	QuantityOnOrder      float64 `json:"quantityOnOrder"`
	QuantityCommitted    float64 `json:"quantityCommitted"`
	QuantityAvailable    float64 `json:"quantityAvailable"`
	Price                float64 `json:"price"`
	Price2               float64 `json:"price2"`
	Price3               float64 `json:"price3"`
	Cost                 float64 `json:"cost"`
	AverageCost          float64 `json:"averageCost"`
	LastCost             float64 `json:"lastCost"`
	UnitOfMeasure        string  `json:"unitOfMeasure"`
	Department           string  `json:"department"`
	VendorID             string  `json:"vendorId"`
	VendorName           string  `json:"vendorName"`
	VendorPartNumber     string  `json:"vendorPartNumber"`
	ManufacturerPartNum  string  `json:"manufacturerPartNum"`
	TaxFlag              bool    `json:"taxFlag"`
	Active               bool    `json:"active"`
	ReorderLevel         float64 `json:"reorderLevel"`
	ReorderQty           float64 `json:"reorderQty"`
	MinOrderQty          float64 `json:"minOrderQty"`
	MaxOrderQty          float64 `json:"maxOrderQty"`
	LeadTimeDays         int     `json:"leadTimeDays"`
	BinLocation          string  `json:"binLocation"`
	Notes                string  `json:"notes"`
	LastModified         string  `json:"lastModified"`
	LastSaleDate         string  `json:"lastSaleDate"`
	LastReceiptDate      string  `json:"lastReceiptDate"`
	SerializedInventory  bool    `json:"serializedInventory"`
	LotTrackedInventory  bool    `json:"lotTrackedInventory"`
}

// NewInventoryPartResponse creates a response from DME InventoryPart model
func NewInventoryPartResponse(part dme.InventoryPart) InventoryPartResponse {
	return InventoryPartResponse{
		PartNumber:           part.PartNumber,
		Description:          part.Description,
		LocationCode:         part.LocationCode,
		QuantityOnHand:       part.QuantityOnHand,
		QuantityOnOrder:      part.QuantityOnOrder,
		QuantityCommitted:    part.QuantityCommitted,
		QuantityAvailable:    part.QuantityAvailable,
		Price:                part.Price,
		Price2:               part.Price2,
		Price3:               part.Price3,
		Cost:                 part.Cost,
		AverageCost:          part.AverageCost,
		LastCost:             part.LastCost,
		UnitOfMeasure:        part.UnitOfMeasure,
		Department:           part.Department,
		VendorID:             part.VendorID,
		VendorName:           part.VendorName,
		VendorPartNumber:     part.VendorPartNumber,
		ManufacturerPartNum:  part.ManufacturerPartNum,
		TaxFlag:              part.TaxFlag,
		Active:               part.Active,
		ReorderLevel:         part.ReorderLevel,
		ReorderQty:           part.ReorderQty,
		MinOrderQty:          part.MinOrderQty,
		MaxOrderQty:          part.MaxOrderQty,
		LeadTimeDays:         part.LeadTimeDays,
		BinLocation:          part.BinLocation,
		Notes:                part.Notes,
		LastModified:         part.LastModified,
		LastSaleDate:         part.LastSaleDate,
		LastReceiptDate:      part.LastReceiptDate,
		SerializedInventory:  part.SerializedInventory,
		LotTrackedInventory:  part.LotTrackedInventory,
	}
}
