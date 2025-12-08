package dme

// CustomerContract represents a Unit Sales contract returned by DME
type CustomerContract struct {
	ContractID           string              `json:"contractId"`
	ProspectID           string              `json:"prospectId"`
	CustomerID           string              `json:"customerId"`
	IsQuote              bool                `json:"isQuote"`
	Name                 string              `json:"name"`
	SalesMan             string              `json:"salesMan"`
	LocationCode         string              `json:"locationCode"`
	SalesLocation        string              `json:"salesLocation"`
	SalesDate            string              `json:"salesDate"`
	WrittenDate          string              `json:"writtenDate"`
	TaxSchema            string              `json:"taxSchema"`
	ThirdPartyIdentifier string              `json:"thirdPartyIdentifier"`
	TotalPrices          float64             `json:"totalPrices"`
	ContractTotal        float64             `json:"contractTotal"`
	BalanceDue           float64             `json:"balanceDue"`
	TotalDeposits        float64             `json:"totalDeposits"`
	CustomInformation    []CustomInformation `json:"customInformation"`
	Boats                []UnitBoat          `json:"boats"`
	Motors               []interface{}       `json:"motors"`
	Trailers             []interface{}       `json:"trailers"`
	OtherUnits           []interface{}       `json:"otherUnits"`
	TradeItems           []interface{}       `json:"tradeItems"`
}

// UnitBoat represents a boat entry within a Unit Sales contract
type UnitBoat struct {
	HIN                string        `json:"hin"`
	IgnitionKeyNum     string        `json:"ignitionKeyNum"`
	BoatModelInfo      BoatModelInfo `json:"boatModelInfo"`
	ID                 string        `json:"id"`
	Status             string        `json:"status"`
	Type               string        `json:"type"`
	SerialNumber       string        `json:"serialNumber"`
	StockNumber        string        `json:"stockNumber"`
	Registration       string        `json:"registration"`
	Description        string        `json:"description"`
	CurrentPhysicalLoc string        `json:"currentPhysicalLoc"`
	Color              string        `json:"color"`
	ListPrice          float64       `json:"listPrice"`
	Price1             float64       `json:"price1"`
	Price2             float64       `json:"price2"`
	Price3             float64       `json:"price3"`
	Price4             float64       `json:"price4"`
	Price5             float64       `json:"price5"`
	FreightPrice       float64       `json:"freightPrice"`
	RiggingPrice       float64       `json:"riggingPrice"`
	PrepPrice          float64       `json:"prepPrice"`
	TotalCost          float64       `json:"totalCost"`
	OptionCost         float64       `json:"optionCost"`
	UnitCost           float64       `json:"unitCost"`
	FreightCost        float64       `json:"freightCost"`
	PrepCost           float64       `json:"prepCost"`
	RiggingCost        float64       `json:"riggingCost"`
	Pack               float64       `json:"pack"`
	ReceivedDate       string        `json:"receivedDate"`
	ManufacturedDate   string        `json:"manufacturedDate"`
	Comments           string        `json:"comments"`
	CustomInfo         []interface{} `json:"customInfo"`
	Options            []interface{} `json:"options"`
	Accessories        []interface{} `json:"accessories"`
	PreviousOwners     []interface{} `json:"previousOwners"`
}

// BoatModelInfo represents model information for a boat in a contract
type BoatModelInfo struct {
	Length          string            `json:"length"`
	Beam            string            `json:"beam"`
	Draft           string            `json:"draft"`
	Weight          string            `json:"weight"`
	HullType        string            `json:"hullType"`
	HullMaterial    string            `json:"hullMaterial"`
	MotorRating     string            `json:"motorRating"`
	FuelCapacity    string            `json:"fuelCapacity"`
	WaterCapacity   string            `json:"waterCapacity"`
	WasteCapacity   string            `json:"wasteCapacity"`
	SleepCapacity   string            `json:"sleepCapacity"`
	CabinHeadroom   string            `json:"cabinHeadroom"`
	BridgeClearance string            `json:"bridgeClearance"`
	DeadRise        string            `json:"deadRise"`
	LengthOverall   string            `json:"lengthOverall"`
	Class           string            `json:"class"`
	ModelID         string            `json:"modelId"`
	VendorName      string            `json:"vendorName"`
	ModelNumber     string            `json:"modelNumber"`
	Year            string            `json:"year"`
	Desc            string            `json:"desc"`
	UnitCost        float64           `json:"unitCost"`
	FreightCost     float64           `json:"freightCost"`
	PrepCost        float64           `json:"prepCost"`
	RiggingCost     float64           `json:"riggingcost"`
	Pack            float64           `json:"pack"`
	ListPrice       float64           `json:"listPrice"`
	Price1          float64           `json:"price1"`
	Price2          float64           `json:"price2"`
	Price3          float64           `json:"price3"`
	Price4          float64           `json:"price4"`
	Price5          float64           `json:"price5"`
	ModelType       string            `json:"modelType"`
	Options         []BoatModelOption `json:"options"`
}

// BoatModelOption represents an option on a boat model
type BoatModelOption struct {
	OptionCode string `json:"optionCode"`
	Desc       string `json:"desc"`
	Price      string `json:"price"`
	GroupCode  string `json:"groupCode"`
	ModelType  string `json:"modelType"`
	ModelID    string `json:"modelId"`
	ModelCount string `json:"modelCount"`
}
