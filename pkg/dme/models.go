package dme

//
// CUSTOMER MODELS
//

// CategoryCode represents a category code
type CategoryCode struct {
	ID   string `json:"id"`
	Desc string `json:"desc"`
}

// CustomInformation represents custom field information
type CustomInformation struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
}

// CustomerShort represents short customer information
type CustomerShort struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	CompanyName string `json:"companyName"`
}

// CustomerListShort represents a paginated list of customers
type CustomerListShort struct {
	Content     []CustomerShort `json:"content"`
	CurrentPage int             `json:"currentPage"`
	MaxPages    int             `json:"maxPages"`
	PageSize    int             `json:"pageSize"`
}

// Customer Search
type CustomerSearch struct {
	CustomerID  string `json:"customerID"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	CompanyName string `json:"companyName"`
	City        string `json:"city"`
	State       string `json:"state"`
	Zip         string `json:"zip"`
	ProspectID  string `json:"prospectId"`
	Inactive    bool   `json:"inactive"`
}

// Attachment represents customer attachment information
type Attachment struct {
	FileName    string  `json:"fileName"`
	Description string  `json:"description"`
	S3Path      string  `json:"s3Path"`
	FileType    *string `json:"fileType"`
	FromDMWeb   *bool   `json:"fromDMWeb"`
}

// Customer represents detailed customer information
type Customer struct {
	ID                          string              `json:"id"`
	Name                        string              `json:"name"`
	FirstName                   string              `json:"firstName"`
	LastName                    string              `json:"lastName"`
	Address1                    string              `json:"address1"`
	Address2                    string              `json:"address2"`
	Address3                    string              `json:"address3"`
	City                        string              `json:"city"`
	State                       string              `json:"state"`
	Zip                         string              `json:"zip"`
	Country                     string              `json:"country"`
	Phone                       string              `json:"phone"`
	AltFirstName                string              `json:"altFirstName"`
	AltLastName                 string              `json:"altLastName"`
	AltAddress1                 string              `json:"altAddress1"`
	AltAddress2                 string              `json:"altAddress2"`
	AltAddress3                 string              `json:"altAddress3"`
	AltCity                     string              `json:"altCity"`
	AltState                    string              `json:"altState"`
	AltZip                      string              `json:"altZip"`
	AltCountry                  string              `json:"altCountry"`
	AltPhone                    string              `json:"altPhone"`
	UseAltAddress               bool                `json:"useAltAddress"`
	CreditLimit                 float64             `json:"creditLimit"`
	WorkPhone                   string              `json:"workPhone"`
	CellPhone                   string              `json:"cellPhone"`
	InactiveDate                string              `json:"inactiveDate"`
	Email                       string              `json:"email"`
	Balance                     float64             `json:"balance"`
	PriceColumn                 string              `json:"priceColumn"`
	SendMassEmail               bool                `json:"sendMassEmail"`
	Discount                    float64             `json:"discount"`
	EmergencyContact            string              `json:"emergencyContact"`
	EmergencyPhone              string              `json:"emergencyPhone"`
	LastModified                string              `json:"lastModified"`
	Inactive                    bool                `json:"inactive"`
	AllowTransactions           bool                `json:"allowTransactions"`
	CategoryCodes               []CategoryCode      `json:"categoryCodes"`
	Boats                       []Boat              `json:"boats"`
	Invoices                    []InvoiceDetailed   `json:"invoices"`
	TaxFlag                     bool                `json:"taxFlag"`
	TaxSchema                   string              `json:"taxSchema"`
	TaxID                       string              `json:"taxId"`
	TaxIDState                  string              `json:"taxIdState"`
	WebID                       string              `json:"webId"`
	WebPassword                 string              `json:"webPassword"`
	PORequired                  bool                `json:"poRequired"`
	AllowBackOrders             bool                `json:"allowBackOrders"`
	Comments                    string              `json:"comments"`
	IntegrationID               string              `json:"integrationId"`
	CompanyName                 string              `json:"companyName"`
	ProspectID                  string              `json:"prospectId"`
	PaymentTermsCode            string              `json:"paymentTermsCode"`
	PaymentTermsCodeDescription string              `json:"paymentTermsCodeDescription"`
	ShipmentMethod              string              `json:"shipmentMethod"`
	ShipmentMethodDescription   string              `json:"shipmentMethodDescription"`
	NoCcSurcharge               bool                `json:"noCcSurcharge"`
	CustomInformation           []CustomInformation `json:"customInformation"`
	WaitListEntries             []WaitListEntry     `json:"waitListEntries"`
	Attachments                 []Attachment        `json:"attachments"`
}

type CustomerList struct {
	Content     []Customer `json:"content"`
	CurrentPage int        `json:"currentPage"`
	MaxPages    int        `json:"maxPages"`
	PageSize    int        `json:"pageSize"`
}

// CustomerUpdate represents updatable customer information
type CustomerUpdate struct {
	ID                        string              `json:"id"`
	Name                      string              `json:"name"`
	FirstName                 string              `json:"firstName"`
	LastName                  string              `json:"lastName"`
	Email                     string              `json:"email"`
	Address1                  string              `json:"address1"`
	Address2                  string              `json:"address2"`
	Address3                  string              `json:"address3"`
	City                      string              `json:"city"`
	State                     string              `json:"state"`
	Zip                       string              `json:"zip"`
	Country                   string              `json:"country"`
	Phone                     string              `json:"phone"`
	AltFirstName              string              `json:"altFirstName"`
	AltLastName               string              `json:"altLastName"`
	AltAddress1               string              `json:"altAddress1"`
	AltAddress2               string              `json:"altAddress2"`
	AltAddress3               string              `json:"altAddress3"`
	AltCity                   string              `json:"altCity"`
	AltState                  string              `json:"altState"`
	AltZip                    string              `json:"altZip"`
	AltCountry                string              `json:"altCountry"`
	AltPhone                  string              `json:"altPhone"`
	UseAltAddress             bool                `json:"useAltAddress"`
	WorkPhone                 string              `json:"workPhone"`
	CellPhone                 string              `json:"cellPhone"`
	EmergencyContact          string              `json:"emergencyContact"`
	EmergencyPhone            string              `json:"emergencyPhone"`
	CompanyName               string              `json:"companyName"`
	ShipmentMethod            string              `json:"shipmentMethod"`
	ShipmentMethodDescription string              `json:"shipmentMethodDescription"`
	CustomInformation         []CustomInformation `json:"customInformation"`
	Attachments               []Attachment        `json:"attachments"`
	Inactive                  bool                `json:"inactive"`
}

// Customer Create
type CustomerCreate struct {
	Name                      string              `json:"name"`
	FirstName                 string              `json:"firstName"`
	LastName                  string              `json:"lastName"`
	Email                     string              `json:"email"`
	Address1                  string              `json:"address1"`
	Address2                  string              `json:"address2"`
	Address3                  string              `json:"address3"`
	City                      string              `json:"city"`
	State                     string              `json:"state"`
	Zip                       string              `json:"zip"`
	Country                   string              `json:"country"`
	Phone                     string              `json:"phone"`
	AltFirstName              string              `json:"altFirstName"`
	AltLastName               string              `json:"altLastName"`
	AltAddress1               string              `json:"altAddress1"`
	AltAddress2               string              `json:"altAddress2"`
	AltAddress3               string              `json:"altAddress3"`
	AltCity                   string              `json:"altCity"`
	AltState                  string              `json:"altState"`
	AltZip                    string              `json:"altZip"`
	AltCountry                string              `json:"altCountry"`
	AltPhone                  string              `json:"altPhone"`
	UseAltAddress             bool                `json:"useAltAddress"`
	WorkPhone                 string              `json:"workPhone"`
	CellPhone                 string              `json:"cellPhone"`
	EmergencyContact          string              `json:"emergencyContact"`
	EmergencyPhone            string              `json:"emergencyPhone"`
	CompanyName               string              `json:"companyName"`
	ShipmentMethod            string              `json:"shipmentMethod"`
	ShipmentMethodDescription string              `json:"shipmentMethodDescription"`
	CustomInformation         []CustomInformation `json:"customInformation"`
	Attachments               []Attachment        `json:"attachments"`
}

type CustomerCreateUpdateResponse struct {
	CustomerID string `json:"customerID"`
}

// WaitListEntry represents wait list entry information
type WaitListEntry struct {
	ID           string `json:"id"`
	WaitListName string `json:"waitListName"`
	Description  string `json:"description"`
	EntryNumber  string `json:"entryNumber"`
	EntryDate    string `json:"entryDate"`
	CustomerID   string `json:"customerId"`
	BoatID       string `json:"boatId"`
	BoatName     string `json:"boatName"`
}

// CustomerMinimal represents a minimal customer information set for list by page
// Only includes fields required by the new API contract
type CustomerMinimal struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Phone      string `json:"phone"`
	WorkPhone  string `json:"workPhone"`
	CellPhone  string `json:"cellPhone"`
	Email      string `json:"emailAddress"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	Address3   string `json:"address3"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
	Inactive   bool   `json:"inactive"`
}

type CustomerListMinimal struct {
	Content     []CustomerMinimal `json:"content"`
	CurrentPage int               `json:"currentPage"`
	MaxPages    int               `json:"maxPages"`
	PageSize    int               `json:"pageSize"`
}

// CustomerMinimalWithCategoryCodes represents minimal customer info with category codes
// Used for esignature and mass notification features
type CustomerMinimalWithCategoryCodes struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	Inactive      bool           `json:"inactive"`
	CategoryCodes []CategoryCode `json:"categoryCodes"`
}

// CustomerWithCategoryCodesPage represents a paginated list of customers with category codes
type CustomerWithCategoryCodesPage struct {
	Content     []CustomerMinimalWithCategoryCodes `json:"content"`
	CurrentPage int                                `json:"currentPage"`
	MaxPages    int                                `json:"maxPages"`
	PageSize    int                                `json:"pageSize"`
	ListName    string                             `json:"listName"`
}

//
// BOAT MODELS
//

// BoatList represents a paginated list of boats
type BoatList struct {
	Content     []Boat `json:"content"`
	CurrentPage int    `json:"currentPage"`
	MaxPages    int    `json:"maxPages"`
	PageSize    int    `json:"pageSize"`
	ListName    string `json:"listName"`
}

// BoatMinimal represents a minimal boat information set for list by page
// Only includes fields required by the new API contract
type BoatMinimal struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	OwnerID      string `json:"ownerId"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Slip         Slip   `json:"slip"`
	LastModified string `json:"lastModified"`
}

type BoatListMinimal struct {
	Content     []BoatMinimal `json:"content"`
	CurrentPage int           `json:"currentPage"`
	MaxPages    int           `json:"maxPages"`
	PageSize    int           `json:"pageSize"`
}

// Motor represents motor information
type Motor struct {
	Number        int      `json:"number"`
	Year          string   `json:"year"`
	Make          string   `json:"make"`
	Model         string   `json:"model"`
	Serial        string   `json:"serial"`
	TransomID     string   `json:"transomId"`
	Drive         string   `json:"drive"`
	Size          string   `json:"size"`
	Hours         *float64 `json:"hours,omitempty"`
	HP            int      `json:"hp"`
	FuelType      string   `json:"fuelType"`
	FuelTypeDesc  string   `json:"fuelTypeDesc"`
	Condition     string   `json:"condition"`
	ConditionDesc string   `json:"conditionDesc"`
	Type          string   `json:"type"`
	TypeDesc      string   `json:"typeDesc"`
}

// Drive represents drive information
type Drive struct {
	Number        int    `json:"number"`
	Year          string `json:"year"`
	Make          string `json:"make"`
	Model         string `json:"model"`
	Serial        string `json:"serial"`
	Type          string `json:"type"`
	TypeDesc      string `json:"typeDesc"`
	GearRatio     string `json:"gearRatio"`
	Condition     string `json:"condition"`
	ConditionDesc string `json:"conditionDesc"`
}

// Generator represents generator information
type Generator struct {
	Number        int    `json:"number"`
	Year          string `json:"year"`
	Make          string `json:"make"`
	Model         string `json:"model"`
	Serial        string `json:"serial"`
	Kilowatts     string `json:"kilowatts"`
	Hours         *int   `json:"hours,omitempty"`
	FuelType      string `json:"fuelType"`
	FuelTypeDesc  string `json:"fuelTypeDesc"`
	Condition     string `json:"condition"`
	ConditionDesc string `json:"conditionDesc"`
}

// Rate represents rate information
type Rate struct {
	Rate      float64 `json:"rate"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
}

// BillingCode represents billing code information
type BillingCode struct {
	ID                    string  `json:"id"`
	Description           string  `json:"description"`
	Department            string  `json:"department"`
	DepartmentDesc        string  `json:"departmentDesc"`
	Cycle                 string  `json:"cycle"`
	PerFoot               bool    `json:"perFoot"`
	ProRated              bool    `json:"proRated"`
	SlipBoatOrLongest     string  `json:"Slip_Boat_Or_Longest"`
	LengthAreaOrCubicFeet string  `json:"Length_Area_Or_CubicFeet"`
	LoaLwlOrSpar          string  `json:"LOA_LWL_Or_Spar"`
	Rates                 []Rate  `json:"rates"`
	OverrideRate          float64 `json:"overrideRate"`
}

// BoatDescriptionCode represents boat description code information
type BoatDescriptionCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// OperationHistory represents operation history information
type OperationHistory struct {
	WorkOrder        string  `json:"workOrder"`
	Code             string  `json:"code"`
	Type             string  `json:"type"`
	Date             string  `json:"date"`
	OperationCharges float64 `json:"operationCharges"`
	Description      string  `json:"description"`
}

// Slip represents slip information
type Slip struct {
	ID              string `json:"id"`
	Description     string `json:"description"`
	Type            string `json:"type"`
	Location        string `json:"location"`
	Length          string `json:"length"`
	Width           string `json:"width"`
	Draft           string `json:"draft"`
	TieOff          string `json:"tieOff"`
	LastModifedDate string `json:"lastModifedDate"`
	Water           bool   `json:"water"`
	Electric        bool   `json:"electric"`
	Phone           bool   `json:"phone"`
	CableTV         bool   `json:"cableTv"`
	Transient       bool   `json:"transient"`
	Linear          bool   `json:"linear"`
	Unusable        bool   `json:"unusable"`
}

// Boat represents boat detail information
type Boat struct {
	DoNotLaunch          bool                  `json:"doNotLaunch"`
	Motors               []Motor               `json:"motors"`
	Drives               []Drive               `json:"drives"`
	Generators           []Generator           `json:"generators"`
	BillingCodes         []BillingCode         `json:"billingCodes"`
	BoatDescriptionCodes []BoatDescriptionCode `json:"boatDescriptionCodes"`
	CustomInformation    []CustomInformation   `json:"customInformation"`
	OperationsHistory    []OperationHistory    `json:"operationsHistory"`
	ID                   string                `json:"id"`
	OwnerID              string                `json:"ownerId"`
	IntegrationID        string                `json:"integrationId"`
	OwnerIntegrationID   string                `json:"ownerIntegrationId"`
	Name                 string                `json:"name"`
	Registration         string                `json:"registration"`
	Year                 string                `json:"year"`
	Make                 string                `json:"make"`
	Model                string                `json:"model"`
	HIN                  string                `json:"hin"`
	LOA                  string                `json:"loa"`
	LWL                  string                `json:"lwl"`
	Draft                string                `json:"draft"`
	Beam                 string                `json:"beam"`
	Height               string                `json:"height"`
	LastModified         string                `json:"lastModified"`
	Color                string                `json:"color"`
	TrailerMake          string                `json:"trailerMake"`
	TrailerModel         string                `json:"trailerModel"`
	TrailerSerial        string                `json:"trailerSerial"`
	TrailerRegistration  string                `json:"trailerRegistration"`
	TrailerLocation      string                `json:"trailerLocation"`
	SummerSlip           string                `json:"summerSlip"`
	WinterSlip           string                `json:"winterSlip"`
	InsuranceCompany     string                `json:"insuranceCompany"`
	InsuranceExpDate     string                `json:"insuranceExpDate"`
	Comments             string                `json:"comments"`
	ContractStartDate    string                `json:"contractStartDate"`
	ContractEndDate      string                `json:"contractEndDate"`
	TransomType          string                `json:"transomType"`
	TransomTypeDesc      string                `json:"transomTypeDesc"`
	TransomHeight        string                `json:"transomHeight"`
	TransomMaterial      string                `json:"transomMaterial"`
	TransomCondition     string                `json:"transomCondition"`
	Access               string                `json:"access"`
	SlipID               string                `json:"slipId"`
	Slip                 Slip                  `json:"slip"`
	Attachments          []Attachment          `json:"attachments"`
}

// BoatUpdate represents boat update information
type BoatUpdate struct {
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	Registration         string                `json:"registration"`
	Year                 string                `json:"year"`
	Make                 string                `json:"make"`
	Model                string                `json:"model"`
	HIN                  string                `json:"hin"`
	LOA                  string                `json:"loa"`
	LWL                  string                `json:"lwl"`
	Draft                string                `json:"draft"`
	Beam                 string                `json:"beam"`
	Height               string                `json:"height"`
	Color                string                `json:"color"`
	TrailerMake          string                `json:"trailerMake"`
	TrailerModel         string                `json:"trailerModel"`
	TrailerSerial        string                `json:"trailerSerial"`
	TrailerRegistration  string                `json:"trailerRegistration"`
	TrailerLocation      string                `json:"trailerLocation"`
	SummerSlip           string                `json:"summerSlip"`
	WinterSlip           string                `json:"winterSlip"`
	InsuranceCompany     string                `json:"insuranceCompany"`
	InsuranceExpDate     string                `json:"insuranceExpDate"`
	SlipID               string                `json:"slipId"`
	Slip                 Slip                  `json:"slip"`
	Motors               []Motor               `json:"motors"`
	Drives               []Drive               `json:"drives"`
	Generators           []Generator           `json:"generators"`
	DoNotLaunch          bool                  `json:"doNotLaunch"`
	BillingCodes         []BillingCode         `json:"billingCodes"`
	BoatDescriptionCodes []BoatDescriptionCode `json:"boatDescriptionCodes"`
	CustomInformation    []CustomInformation   `json:"customInformation"`
	OperationsHistory    []OperationHistory    `json:"operationsHistory"`
	IntegrationID        string                `json:"integrationId"`
	OwnerIntegrationID   string                `json:"ownerIntegrationId"`
	LastModified         string                `json:"lastModified"`
	Comments             string                `json:"comments"`
	ContractStartDate    string                `json:"contractStartDate"`
	ContractEndDate      string                `json:"contractEndDate"`
	TransomType          string                `json:"transomType"`
	TransomTypeDesc      string                `json:"transomTypeDesc"`
	TransomHeight        string                `json:"transomHeight"`
	TransomMaterial      string                `json:"transomMaterial"`
	TransomCondition     string                `json:"transomCondition"`
	Access               string                `json:"access"`
	Attachments          []Attachment          `json:"attachments"`
}

// BoatCreate represents boat creation information
type BoatCreate struct {
	Name                 string                `json:"name"`
	OwnerID              string                `json:"ownerId"`
	Registration         string                `json:"registration"`
	Year                 string                `json:"year"`
	Make                 string                `json:"make"`
	Model                string                `json:"model"`
	HIN                  string                `json:"hin"`
	LOA                  string                `json:"loa"`
	LWL                  string                `json:"lwl"`
	Draft                string                `json:"draft"`
	Beam                 string                `json:"beam"`
	Height               string                `json:"height"`
	Color                string                `json:"color"`
	TrailerMake          string                `json:"trailerMake"`
	TrailerModel         string                `json:"trailerModel"`
	TrailerSerial        string                `json:"trailerSerial"`
	TrailerRegistration  string                `json:"trailerRegistration"`
	TrailerLocation      string                `json:"trailerLocation"`
	SummerSlip           string                `json:"summerSlip"`
	WinterSlip           string                `json:"winterSlip"`
	InsuranceCompany     string                `json:"insuranceCompany"`
	InsuranceExpDate     string                `json:"insuranceExpDate"`
	SlipID               string                `json:"slipId"`
	Slip                 Slip                  `json:"slip"`
	Motors               []Motor               `json:"motors"`
	Drives               []Drive               `json:"drives"`
	Generators           []Generator           `json:"generators"`
	DoNotLaunch          bool                  `json:"doNotLaunch"`
	BillingCodes         []BillingCode         `json:"billingCodes"`
	BoatDescriptionCodes []BoatDescriptionCode `json:"boatDescriptionCodes"`
	CustomInformation    []CustomInformation   `json:"customInformation"`
	OperationsHistory    []OperationHistory    `json:"operationsHistory"`
	IntegrationID        string                `json:"integrationId"`
	OwnerIntegrationID   string                `json:"ownerIntegrationId"`
	LastModified         string                `json:"lastModified"`
	Comments             string                `json:"comments"`
	ContractStartDate    string                `json:"contractStartDate"`
	ContractEndDate      string                `json:"contractEndDate"`
	TransomType          string                `json:"transomType"`
	TransomTypeDesc      string                `json:"transomTypeDesc"`
	TransomHeight        string                `json:"transomHeight"`
	TransomMaterial      string                `json:"transomMaterial"`
	TransomCondition     string                `json:"transomCondition"`
	Access               string                `json:"access"`
	Attachments          []Attachment          `json:"attachments"`
}

type BoatCreateUpdateResponse struct {
	BoatID string `json:"boatID"`
}

// BoatSearch represents search results for boats
type BoatSearch struct {
	BoatID        string `json:"boatId"`
	Name          string `json:"boatName"`
	OwnerName     string `json:"ownerName"`
	ArrivalDate   string `json:"arrivalDate"`
	DepartureDate string `json:"departureDate"`
	HIN           string `json:"hin"`
}

//
// GENERAL MODELS
//

// Clerk represents a system user/clerk in the DME system
// This struct matches the actual DockMaster API response
type Clerk struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	EmailAddress      string `json:"emailAddress"`
	GroupEmailAddress string `json:"groupEmailAddress"`
	InactiveDate      string `json:"inactiveDate"`
	IsActive          bool   `json:"isActive"`
}

//
// OTHER MODELS
//

// VersionInfo represents the system version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
}

// Installment represents installment information
type Installment struct {
	ID      string  `json:"id"`
	DueDate string  `json:"dueDate"`
	Amount  float64 `json:"amount"`
	Balance float64 `json:"balance"`
}

// InvoiceDetailed represents detailed invoice information
type InvoiceDetailed struct {
	ID               string        `json:"id"`
	CustomerID       string        `json:"customerId"`
	DueDate          string        `json:"dueDate"`
	Description      string        `json:"description"`
	LocationCode     string        `json:"locationCode"`
	UnappliedPayment bool          `json:"unappliedPayment"`
	Installments     []Installment `json:"installments"`
	Amount           float64       `json:"amount"`
	InvoiceAmount    float64       `json:"invoiceAmount"`
	InvoiceBalance   float64       `json:"invoiceBalance"`
}

// LocationDetailed represents detailed location information
type LocationDetailed struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LocationNumber string `json:"locationNumber"`
	Address1       string `json:"address1"`
	Address2       string `json:"address2"`
	Address3       string `json:"address3"`
	City           string `json:"city"`
	State          string `json:"state"`
	ZipCode        string `json:"zipCode"`
	Country        string `json:"country"`
	Phone          string `json:"phone"`
	Fax            string `json:"fax"`
	BillToAddress1 string `json:"billToAddress1"`
	BillToAddress2 string `json:"billToAddress2"`
	BillToAddress3 string `json:"billToAddress3"`
	BillToCity     string `json:"billToCity"`
	BillToState    string `json:"billToState"`
	BillToZip      string `json:"billToZip"`
	BillToCountry  string `json:"billToCountry"`
	BillToPhone    string `json:"billToPhone"`
	BillToFax      string `json:"billToFax"`
	DmPayClientID  string `json:"dmPayClientId"`
}

// WorkOrderShort represents short work order information
type WorkOrderShort struct {
	ID           string `json:"id"`
	Customer     string `json:"customer"`
	OpenDate     string `json:"openDate"`
	Boat         string `json:"boat"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	LocationCode string `json:"locationCode"`
}

// Operation represents operation information
type Operation struct {
	ID                     string       `json:"id"`
	Opcode                 string       `json:"opcode"`
	OpcodeDesc             string       `json:"opcodeDesc"`
	Status                 string       `json:"status"`
	Type                   string       `json:"type"`
	Category               string       `json:"category"`
	TotalCharges           float64      `json:"totalCharges"`
	TotalParts             float64      `json:"totalParts"`
	TotalLabor             float64      `json:"totalLabor"`
	TotalLaborHours        float64      `json:"totalLaborHours"`
	TotalFreight           float64      `json:"totalFreight"`
	TotalEquipment         float64      `json:"totalEquipment"`
	TotalSublet            float64      `json:"totalSublet"`
	TotalMileage           float64      `json:"totalMileage"`
	TotalMiscSupply        float64      `json:"totalMiscSupply"`
	TotalBillCodes         float64      `json:"totalBillCodes"`
	LaborBilled            float64      `json:"laborBilled"`
	TotalToComplete        float64      `json:"totalToComplete"`
	EstStartDate           string       `json:"estStartDate"`
	EstCompleteDate        string       `json:"estCompleteDate"`
	ReqCompDate            string       `json:"reqCompDate"`
	StandardHours          float64      `json:"standardHours"`
	LongDesc               string       `json:"longDesc"`
	TechDesc               string       `json:"techDesc"`
	ManagerComments        string       `json:"managerComments"`
	EstimatedCharges       float64      `json:"estimatedCharges"`
	EstimatedParts         float64      `json:"estimatedParts"`
	EstimatedLabor         float64      `json:"estimatedLabor"`
	EstimatedLaborHours    float64      `json:"estimatedLaborHours"`
	EstimatedFreight       float64      `json:"estimatedFreight"`
	EstimatedEquipment     float64      `json:"estimatedEquipment"`
	EstimatedSublet        float64      `json:"estimatedSublet"`
	EstimatedMileage       float64      `json:"estimatedMileage"`
	EstimatedMiscSupply    float64      `json:"estimatedMiscSupply"`
	EstimatedBillCodes     float64      `json:"estimatedBillCodes"`
	IsOpcodeApproved       bool         `json:"isOpcodeApproved"`
	FlatRateAmount         float64      `json:"flatRateAmount"`
	FlatRatePerFootRate    float64      `json:"flatRatePerFootRate"`
	FlatRatePerFootMethod  string       `json:"flatRatePerFootMethod"`
	ForecastedPartsCharges float64      `json:"forecastedPartsCharges"`
	ForecastedLaborCharges float64      `json:"forecastedLaborCharges"`
	ForecastedLaborHours   float64      `json:"forecastedLaborHours"`
	FlagLaborFinished      bool         `json:"flagLaborFinished"`
	Attachments            []Attachment `json:"attachments"`
}

// BillingData represents billing data information
type BillingData struct {
	BillingDate        string  `json:"billingDate"`
	AmountBilled       float64 `json:"amountBilled"`
	EnvironmentCharges float64 `json:"environmentCharges"`
	OtherCharges       float64 `json:"otherCharges"`
	SalesTax           float64 `json:"salesTax"`
}

// WorkOrderFull represents full work order information
type WorkOrderFull struct {
	ID                   string        `json:"id"`
	CustomerID           string        `json:"customerID"`
	CustomerName         string        `json:"customerName"`
	CreationDate         string        `json:"creationDate"`
	ClerkID              string        `json:"clerkId"`
	Type                 string        `json:"type"`
	StartDate            string        `json:"startDate"`
	Status               string        `json:"status"`
	TaxSchema            string        `json:"taxSchema"`
	BoatID               string        `json:"boatId"`
	BoatName             string        `json:"boatName"`
	BoatYear             string        `json:"boatYear"`
	BoatMake             string        `json:"boatMake"`
	BoatModel            string        `json:"boatModel"`
	BoatLength           string        `json:"boatLength"`
	EstCompDate          string        `json:"estCompDate"`
	TotalParts           float64       `json:"totalParts"`
	TotalLabor           float64       `json:"totalLabor"`
	TotalLaborHours      float64       `json:"totalLaborHours"`
	TotalFreight         float64       `json:"totalFreight"`
	TotalEquipment       float64       `json:"totalEquipment"`
	TotalSublet          float64       `json:"totalSublet"`
	TotalMileage         float64       `json:"totalMileage"`
	TotalMiscSupply      float64       `json:"totalMiscSupply"`
	TotalBillCodes       float64       `json:"totalBillCodes"`
	TotalWOCharges       float64       `json:"totalWOCharges"`
	TotalPartsCost       float64       `json:"totalPartsCost"`
	TotalLaborCost       float64       `json:"totalLaborCost"`
	TotalSubletCost      float64       `json:"totalSubletCost"`
	TotalFreightCost     float64       `json:"totalFreightCost"`
	TotalForecastedParts float64       `json:"totalForecastedParts"`
	TotalForecastedLabor float64       `json:"totalForecastedLabor"`
	TotalForecastedHours float64       `json:"totalForecastedHours"`
	Category             string        `json:"category"`
	LocationCode         string        `json:"locationCode"`
	EstStartDate         string        `json:"estStartDate"`
	PromisedDate         string        `json:"promisedDate"`
	LastModDate          string        `json:"lastModDate"`
	LastModTime          string        `json:"lastModTime"`
	Comments             string        `json:"comments"`
	IsEstimate           bool          `json:"isEstimate"`
	RiggingID            string        `json:"riggingId"`
	RiggingType          string        `json:"riggingType"`
	Title                string        `json:"title"`
	Operations           []Operation   `json:"operations"`
	BillingData          []BillingData `json:"billingData"`
	Attachments          []Attachment  `json:"attachments"`
}

// Location represents location information
type Location struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LocationNumber string `json:"locationNumber"`
	Address1       string `json:"address1"`
	Address2       string `json:"address2"`
	Address3       string `json:"address3"`
	City           string `json:"city"`
	State          string `json:"state"`
	ZipCode        string `json:"zipCode"`
	Country        string `json:"country"`
	Phone          string `json:"phone"`
	Fax            string `json:"fax"`
	BillToAddress1 string `json:"billToAddress1"`
	BillToAddress2 string `json:"billToAddress2"`
	BillToAddress3 string `json:"billToAddress3"`
	BillToCity     string `json:"billToCity"`
	BillToState    string `json:"billToState"`
	BillToZip      string `json:"billToZip"`
	BillToCountry  string `json:"billToCountry"`
	BillToPhone    string `json:"billToPhone"`
	BillToFax      string `json:"billToFax"`
	DMPayClientID  string `json:"dmPayClientId"`
}

// Department represents a department in the system
type Department struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	InternalCOS string `json:"internalCOS"`
	RetailCOS   string `json:"retailCOS"`
}

// WorkOrderList represents a paginated list of work orders
type WorkOrderList struct {
	Content     []WorkOrder `json:"content"`
	CurrentPage int         `json:"currentPage"`
	MaxPages    int         `json:"maxPages"`
	PageSize    int         `json:"pageSize"`
}

// WorkOrder represents a full work order
type WorkOrder = WorkOrderFull

// WorkOrderSearch represents a work order search result
type WorkOrderSearch struct {
	ID           string `json:"id"`
	Customer     string `json:"customer"`
	Boat         string `json:"boat"`
	OpenDate     string `json:"openDate"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	LocationCode string `json:"locationCode"`
}

// WorkOrderListShort represents a paginated list of work order short information
type WorkOrderListShort struct {
	Content     []WorkOrderShort `json:"content"`
	CurrentPage int              `json:"currentPage"`
	MaxPages    int              `json:"maxPages"`
	PageSize    int              `json:"pageSize"`
}

// WorkOrderCreateResponse represents the response from creating or updating a work order
type WorkOrderCreateResponse struct {
	WoId       string   `json:"woId"`
	Operations []string `json:"operations"`
	Result     string   `json:"result"`
}

// WorkOrderOperation represents a work order operation retrieved from the API
type WorkOrderOperation struct {
	Opcode                 string  `json:"opcode"`
	Desc                   string  `json:"desc"`
	LongDesc               string  `json:"longDesc"`
	TechDesc               string  `json:"techDesc"`
	CategoryCode           string  `json:"categoryCode"`
	EstimatedParts         float64 `json:"estimatedParts"`
	EstimatedLabor         float64 `json:"estimatedLabor"`
	EstimatedLaborHours    float64 `json:"estimatedLaborHours"`
	EstimatedFreight       float64 `json:"estimatedFreight"`
	EstimatedEquipment     float64 `json:"estimatedEquipment"`
	EstimatedSublet        float64 `json:"estimatedSublet"`
	EstimatedMileage       float64 `json:"estimatedMileage"`
	EstimatedMiscSupply    float64 `json:"estimatedMiscSupply"`
	EstimatedBillCodes     float64 `json:"estimatedBillCodes"`
	Approved               bool    `json:"approved"`
	FlatRateAmount         float64 `json:"flatRateAmount"`
	FlatRatePerFootRate    float64 `json:"flatRatePerFootRate"`
	FlatRatePerFootMethod  string  `json:"flatRatePerFootMethod"`
	LaborFinished          bool    `json:"laborFinished"`
	StandardHours          float64 `json:"standardHours"`
	EstCompDate            string  `json:"estCompDate"`
	EstStartDate           string  `json:"estStartDate"`
	CustPromiseDate        string  `json:"custPromiseDate"`
	ReqCompDate            string  `json:"reqCompDate"`
	ForecastedPartsCharges float64 `json:"forecastedPartsCharges"`
	ForecastedLaborCharges float64 `json:"forecastedLaborCharges"`
	ForecastedLaborHours   float64 `json:"forecastedLaborHours"`
	ManagerComments        string  `json:"managerComments"`
}

// PaginationRequest represents pagination parameters for POST endpoints
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// OperationsListResponse represents paginated list of operations
type OperationsListResponse struct {
	CurrentPage int                  `json:"currentPage"`
	MaxPages    int                  `json:"maxPages"`
	PageSize    int                  `json:"pageSize"`
	ListName    string               `json:"listName"`
	Content     []WorkOrderOperation `json:"content"`
}

// WorkOrderDetailPartEntry represents a part entry in a work order
type WorkOrderDetailPartEntry struct {
	TechID       string  `json:"techId"`
	WorkOrderID  string  `json:"workOrderId"`
	OpCode       string  `json:"opCode"`
	Department   string  `json:"department"`
	Date         string  `json:"date"`
	PartID       string  `json:"partId"`
	Description  string  `json:"description"`
	LocationCode string  `json:"locationCode"`
	Quantity     float64 `json:"quantity"`
	Price        float64 `json:"price"`
	Cost         float64 `json:"cost"`
	Discount     float64 `json:"discount"`
	Comments     string  `json:"comments"`
	UID          string  `json:"uid"`
}

// WorkOrderPartDetail represents detailed part information for a work order operation
type WorkOrderPartDetail struct {
	UID                string   `json:"uid"`
	WorkOrderDetailID  string   `json:"work_order_detail_id"`
	WorkOrderID        string   `json:"work_order_id"`
	OperationCode      string   `json:"operation_code"`
	Department         string   `json:"department"`
	LineNumber         *int     `json:"line_number"`
	Part               string   `json:"part"`
	PartQty            *float64 `json:"part_qty"`
	PartPrice          *float64 `json:"part_price"`
	PartCost           *float64 `json:"part_cost"`
	PartExtend         *float64 `json:"part_extend"`
	PartDate           *string  `json:"part_date"`
	PartDesc           string   `json:"part_desc"`
	PartTech           string   `json:"part_tech"`
	PartLocCode        string   `json:"part_loc_code"`
	PartLocPos         *int     `json:"part_loc_pos"`
	PartBilledDate     *string  `json:"part_billed_date"`
	SpecialOrder       string   `json:"special_order"`
	PartComment        string   `json:"part_comment"`
	PartDiscount       *float64 `json:"part_discount"`
	PartSerialNum      string   `json:"part_serial_num"`
	PartInvoiceID      string   `json:"part_invoice_id"`
	PickListPrint      *bool    `json:"pick_list_print"`
	PartRiggedAcc      *bool    `json:"part_rigged_acc"`
}

// SubletPurchaseOrder represents a sublet purchase order
type SubletPurchaseOrder struct {
	ID              string  `json:"id"`
	VendorID        string  `json:"vendorId"`
	VendorName      string  `json:"vendorName"`
	Date            string  `json:"date"`
	WorkOrderID     string  `json:"workOrderId"`
	Operation       string  `json:"operation"`
	PartsCost       float64 `json:"partsCost"`
	PartsPrice      float64 `json:"partsPrice"`
	PartsDiscount   float64 `json:"partsDiscount"`
	LaborCost       float64 `json:"laborCost"`
	LaborPrice      float64 `json:"laborPrice"`
	LaborDiscount   float64 `json:"laborDiscount"`
	Comments        string  `json:"comments"`
	DepartmentID    string  `json:"departmentId"`
	DepartmentDesc  string  `json:"departmentDesc"`
	BilledDate      string  `json:"billedDate"`
	ClosedDate      string  `json:"closedDate"`
	LocationCode    string  `json:"locationCode"`
	ShipToName      string  `json:"shipToName"`
	ShipToAddress1  string  `json:"shipToAddress1"`
	ShipToAddress2  string  `json:"shipToAddress2"`
	ShipToAddress3  string  `json:"shipToAddress3"`
	ShipToAddress4  string  `json:"shipToAddress4"`
	BillToName      string  `json:"billToName"`
	BillToAddress1  string  `json:"billToAddress1"`
	BillToAddress2  string  `json:"billToAddress2"`
	BillToAddress3  string  `json:"billToAddress3"`
	BillToAddress4  string  `json:"billToAddress4"`
}

// TimeEntryDetail represents a time entry detail in the list
type TimeEntryDetail struct {
	LaborUID    string      `json:"laborUID"`
	TechID      string      `json:"techId"`
	WorkOrderID string      `json:"workOrderId"`
	DepartmentID string     `json:"departmentId"`
	TotalHours  string      `json:"totalHours"`
	EffHours    string      `json:"effHours"`
	Operations  []Operation `json:"operations"`
}

// TimeEntryListResponse represents paginated list of time entries
type TimeEntryListResponse struct {
	CurrentPage int               `json:"currentPage"`
	MaxPages    int               `json:"maxPages"`
	PageSize    int               `json:"pageSize"`
	ListName    string            `json:"listName"`
	Content     []TimeEntryDetail `json:"content"`
}

// OpCodeListResponse represents a paginated list of operation codes
type OpCodeListResponse struct {
	CurrentPage int                  `json:"currentPage"`
	MaxPages    int                  `json:"maxPages"`
	PageSize    int                  `json:"pageSize"`
	ListName    string               `json:"listName"`
	OpCodes     []WorkOrderOperation `json:"opCodes"`
}

// Technician represents a technician record
type Technician struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Active   bool   `json:"active"`
}

//
// PAYMENT MODELS
//

// CustomerInvoiceInquiry represents the response from the AR/CustomerARInquiry endpoint
type CustomerInvoiceInquiry struct {
	CustomerID      string           `json:"customerId"`
	FirstName       string           `json:"firstName"`
	LastName        string           `json:"lastName"`
	Email           string           `json:"email"`
	HomePhone       string           `json:"homePhone"`
	CellPhone       string           `json:"cellPhone"`
	CompanyName     string           `json:"companyName"`
	CurrentBalance  float64          `json:"currentBalance"`
	AgingCurrent    float64          `json:"agingCurrent"`
	Aging30         float64          `json:"aging30"`
	Aging60         float64          `json:"aging60"`
	Aging90         float64          `json:"aging90"`
	Aging120        float64          `json:"aging120"`
	OpenARInvoices  []OpenARInvoice  `json:"openARInvoices"`
	PendingPayments []PendingPayment `json:"pendingPayments"`
	LastPayment     LastPayment      `json:"lastPayment"`
}

// OpenARInvoice represents an open AR invoice in the customer inquiry
type OpenARInvoice struct {
	ID               string        `json:"id"`
	CustomerID       string        `json:"customerId"`
	Amount           float64       `json:"amount"`
	InvoiceAmount    float64       `json:"invoiceAmount"`
	InvoiceBalance   float64       `json:"invoiceBalance"`
	DueDate          string        `json:"dueDate,omitempty"`
	Description      string        `json:"description"`
	TransactionType  string        `json:"transactionType"`
	InvoiceType      string        `json:"invoiceType"`
	SourceID         string        `json:"sourceId"`
	ScheduleAcct     string        `json:"scheduleAcct"`
	LocationCode     string        `json:"locationCode"`
	UnAppliedPayment bool          `json:"unAppliedPayment"`
	Installments     []Installment `json:"installments"`
	BoatName         string        `json:"boatName,omitempty"`
	WohBillingID     string        `json:"wohBillingId,omitempty"`
	JtglTotal        string        `json:"jtglTotal,omitempty"`
}

// PendingPayment represents a pending payment in the customer inquiry
type PendingPayment struct {
	BatchID         string                 `json:"batchId"`
	PaymentDate     string                 `json:"paymentDate"`
	TotalPaymentAmt float64                `json:"totalPaymentAmt"`
	ReferenceNum    string                 `json:"referenceNum"`
	PaymentDetails  []PendingPaymentDetail `json:"paymentDetails"`
}

// PendingPaymentDetail represents details of a pending payment
type PendingPaymentDetail struct {
	InvoiceID         string  `json:"invoiceId"`
	TotalInvoiceAmt   float64 `json:"totalInvoiceAmt"`
	PaymentAmount     float64 `json:"paymentAmount"`
	InvoiceBalance    float64 `json:"invoiceBalance"`
	PaymentActionCode string  `json:"paymentActionCode"`
	PaymentActionDesc string  `json:"paymentActionDesc"`
}

// LastPayment represents the last payment information in the customer inquiry
type LastPayment struct {
	PaymentDate   string  `json:"paymentDate"`
	PaymentAmount float64 `json:"paymentAmount"`
	PaymentMethod string  `json:"paymentMethod"`
}

// PaymentInitiationResponse represents the response when initiating a payment
type PaymentInitiationResponse struct {
	PaymentSessionID string  `json:"paymentSessionId"`
	DMPayClientID    string  `json:"dmPayClientId"`
	CustomerID       string  `json:"customerId"`
	InvoiceID        string  `json:"invoiceId"`
	Amount           float64 `json:"amount"`
	Status           string  `json:"status"`
	PaymentURL       string  `json:"paymentUrl"`
}

//
// INVENTORY MODELS
//

// FuelInventory represents a fuel inventory record
type FuelInventory struct {
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

// OnlineBillcode represents a billing code available for online use
type OnlineBillcode struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Department  string  `json:"department"`
	Price       float64 `json:"price"`
	TaxFlag     bool    `json:"taxFlag"`
	Active      bool    `json:"active"`
}

// OnlinePart represents an inventory part available for online use
type OnlinePart struct {
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

// PartQtyInfo represents quantity information for a specific part at a location
type PartQtyInfo struct {
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

// PartsKitItem represents an item in a parts kit
type PartsKitItem struct {
	PartNumber  string  `json:"partNumber"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`
	Cost        float64 `json:"cost"`
}

// PartsKit represents a parts kit
type PartsKit struct {
	ID           string         `json:"id"`
	Description  string         `json:"description"`
	Active       bool           `json:"active"`
	TotalPrice   float64        `json:"totalPrice"`
	TotalCost    float64        `json:"totalCost"`
	Items        []PartsKitItem `json:"items"`
	LastModified string         `json:"lastModified"`
}

// PurchaseOrderLine represents a line item in a purchase order
type PurchaseOrderLine struct {
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

// PurchaseOrder represents a purchase order
type PurchaseOrder struct {
	ID              string              `json:"id"`
	PONumber        string              `json:"poNumber"`
	VendorID        string              `json:"vendorId"`
	VendorName      string              `json:"vendorName"`
	OrderDate       string              `json:"orderDate"`
	ExpectedDate    string              `json:"expectedDate"`
	ReceivedDate    string              `json:"receivedDate"`
	Status          string              `json:"status"`
	LocationCode    string              `json:"locationCode"`
	TotalAmount     float64             `json:"totalAmount"`
	TotalReceived   float64             `json:"totalReceived"`
	Comments        string              `json:"comments"`
	Lines           []PurchaseOrderLine `json:"lines"`
	CreatedBy       string              `json:"createdBy"`
	LastModifiedBy  string              `json:"lastModifiedBy"`
	LastModified    string              `json:"lastModified"`
}

// PurchaseOrdersList represents a list of purchase orders
type PurchaseOrdersList struct {
	Content     []PurchaseOrder `json:"content"`
	CurrentPage int             `json:"currentPage"`
	MaxPages    int             `json:"maxPages"`
	PageSize    int             `json:"pageSize"`
}

// SpecialOrder represents a special order
type SpecialOrder struct {
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

// InventorySearchResult represents a search result from inventory
type InventorySearchResult struct {
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

// InventoryPart represents a full inventory part record from Retrieve endpoint
type InventoryPart struct {
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

// RetrieveInventoryQuery represents the query parameters for retrieving inventory
type RetrieveInventoryQuery struct {
	LocationCode     string   `json:"locationCode,omitempty"`
	LastModifiedDate string   `json:"lastModifiedDate,omitempty"`
	OnHandOnly       bool     `json:"onHandOnly,omitempty"`
	VendorID         string   `json:"vendorId,omitempty"`
	ItemIds          []string `json:"itemIds,omitempty"`
}

//
// LABOR DETAIL MODELS
//

// LaborEntry represents a labor entry for an estimate or work order
type LaborEntry struct {
	TechID            string  `json:"techId"`
	TechName          string  `json:"techName"`
	WorkOrderID       string  `json:"workOrderId"`
	OpCode            string  `json:"opCode"`
	Department        string  `json:"department"`
	DepartmentDesc    string  `json:"departmentDesc"`
	Date              string  `json:"date"`
	StartTime         string  `json:"startTime"`
	StopTime          string  `json:"stopTime"`
	Hours             float64 `json:"hours"`
	BilledHours       float64 `json:"billedHours"`
	LaborRate         float64 `json:"laborRate"`
	LaborPrice        float64 `json:"laborPrice"`
	LaborCost         float64 `json:"laborCost"`
	Comments          string  `json:"comments"`
	IsApproved        bool    `json:"isApproved"`
	FlagLaborFinished bool    `json:"flagLaborFinished"`
	UID               string  `json:"uid"`
}

// WorkOrderLaborDetail represents detailed labor information for a work order
type WorkOrderLaborDetail struct {
	WorkOrderID    string       `json:"workOrderId"`
	OpCode         string       `json:"opCode"`
	OpDesc         string       `json:"opDesc"`
	TotalHours     float64      `json:"totalHours"`
	TotalBilled    float64      `json:"totalBilled"`
	TotalCost      float64      `json:"totalCost"`
	LaborEntries   []LaborEntry `json:"laborEntries"`
}

// WorkOrderLaborDetailRecord represents comprehensive individual labor detail record (from DockMaster_API WorkOrderLaborDetail endpoint)
type WorkOrderLaborDetailRecord struct {
	WorkOrderDetailID    string   `json:"work_order_detail_id"`
	WorkOrderID          string   `json:"work_order_id"`
	OperationCode        string   `json:"operation_code"`
	Department           string   `json:"department"`
	LineNumber           *int     `json:"line_number"`
	LaborDate            *string  `json:"labor_date"`
	LaborPriceRate       *float64 `json:"labor_price_rate"`
	LaborCostRate        *float64 `json:"labor_cost_rate"`
	LaborEmployee        string   `json:"labor_employee"`
	LaborStartTime       string   `json:"labor_start_time"`
	LaborStopTime        string   `json:"labor_stop_time"`
	LaborPriceExtend     *float64 `json:"labor_price_extend"`
	LaborHours           *float64 `json:"labor_hours"`
	LaborBilledDate      *string  `json:"labor_billed_date"`
	LaborTechComments    string   `json:"labor_tech_comments"`
	LaborCostExtension   *float64 `json:"labor_cost_extension"`
	LaborPriceFactor     string   `json:"labor_price_factor"`
	LaborDiscount        *float64 `json:"labor_discount"`
	LaborInvoiceID       string   `json:"labor_invoice_id"`
	LaborCostFactor      string   `json:"labor_cost_factor"`
	LaborEffectiveHours  *float64 `json:"labor_effective_hours"`
	UID                  *string  `json:"labor_uid"`
}