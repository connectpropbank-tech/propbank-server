package models

import (
	"time"
)

// SpouseInfo represents spouse information for married tenants
type SpouseInfo struct {
	FirstName        string `json:"firstName" firestore:"firstName"`
	LastName         string `json:"lastName" firestore:"lastName"`
	Email            string `json:"email" firestore:"email"`
	Phone            string `json:"phone" firestore:"phone"`
	EmploymentStatus string `json:"employmentStatus" firestore:"employmentStatus"`
	Employer         string `json:"employer" firestore:"employer"`
	Notes            string `json:"notes" firestore:"notes"`
}

// FurnishedItem represents a furnished item with quantity
type FurnishedItem struct {
	ID       string `json:"id" firestore:"id"`
	Name     string `json:"name" firestore:"name"`
	Checked  bool   `json:"checked" firestore:"checked"`
	Quantity int    `json:"quantity" firestore:"quantity"`
	Category string `json:"category" firestore:"category"`
}

// TenantInfo represents tenant information stored in property document
type TenantInfo struct {
	ID string `json:"id" firestore:"id"`

	// Personal Information
	FirstName        string `json:"firstName" firestore:"firstName"`
	LastName         string `json:"lastName" firestore:"lastName"`
	Email            string `json:"email" firestore:"email"`
	Phone            string `json:"phone" firestore:"phone"`
	EmergencyContact string `json:"emergencyContact" firestore:"emergencyContact"`
	UserUID          string `json:"userUID,omitempty" firestore:"userUID,omitempty"` // Map to platform user if found

	// Marital Status
	IsMarried bool        `json:"isMarried" firestore:"isMarried"`
	Spouse    *SpouseInfo `json:"spouse,omitempty" firestore:"spouse,omitempty"`

	// Lease Information
	LeaseStartDate  string `json:"leaseStartDate" firestore:"leaseStartDate"`
	LeaseEndDate    string `json:"leaseEndDate" firestore:"leaseEndDate"`
	MonthlyRent     string `json:"monthlyRent" firestore:"monthlyRent"`
	SecurityDeposit string `json:"securityDeposit" firestore:"securityDeposit"`

	// Payment Details
	PaymentDueDate       string `json:"paymentDueDate" firestore:"paymentDueDate"`
	EscalationPercentage string `json:"escalationPercentage" firestore:"escalationPercentage"`
	EscalationAmount     string `json:"escalationAmount" firestore:"escalationAmount"`

	// Background Information
	PreviousAddress  string `json:"previousAddress" firestore:"previousAddress"`
	EmploymentStatus string `json:"employmentStatus" firestore:"employmentStatus"`
	Employer         string `json:"employer" firestore:"employer"`
	MonthlyIncome    string `json:"monthlyIncome" firestore:"monthlyIncome"`

	// Additional Information
	Notes                          string    `json:"notes" firestore:"notes"`
	IsActive                       bool      `json:"isActive" firestore:"isActive"`
	NoticePeriod                   string    `json:"noticePeriod" firestore:"noticePeriod"`
	LastRentPaymentReminderSentAt  time.Time `json:"lastRentPaymentReminderSentAt" firestore:"lastRentPaymentReminderSentAt"`
	LastNoticePeriodReminderSentAt time.Time `json:"lastNoticePeriodReminderSentAt" firestore:"lastNoticePeriodReminderSentAt"`

	// System fields
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// BuyerInfo represents buyer information stored in property document
type BuyerInfo struct {
	ID string `json:"id" firestore:"id"`

	// Personal Information
	FirstName        string `json:"firstName" firestore:"firstName"`
	LastName         string `json:"lastName" firestore:"lastName"`
	Email            string `json:"email" firestore:"email"`
	Phone            string `json:"phone" firestore:"phone"`
	EmergencyContact string `json:"emergencyContact" firestore:"emergencyContact"`

	// Purchase Information
	OfferAmount       string `json:"offerAmount" firestore:"offerAmount"`
	FinancingType     string `json:"financingType" firestore:"financingType"` // cash, loan, etc.
	PreApprovalAmount string `json:"preApprovalAmount" firestore:"preApprovalAmount"`
	ClosingDate       string `json:"closingDate" firestore:"closingDate"`

	// Background Information
	CurrentAddress   string `json:"currentAddress" firestore:"currentAddress"`
	EmploymentStatus string `json:"employmentStatus" firestore:"employmentStatus"`
	Employer         string `json:"employer" firestore:"employer"`
	AnnualIncome     string `json:"annualIncome" firestore:"annualIncome"`

	// Agent Information (if any)
	AgentName  string `json:"agentName" firestore:"agentName"`
	AgentPhone string `json:"agentPhone" firestore:"agentPhone"`

	// Additional Information
	Notes    string `json:"notes" firestore:"notes"`
	IsActive bool   `json:"isActive" firestore:"isActive"`

	// System fields
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type Property struct {
	ID string `json:"id" firestore:"id"`

	// Basic Property Details
	Title         string `json:"title" firestore:"title"`
	PropertyType  string `json:"propertyType" firestore:"propertyType"`
	Configuration string `json:"configuration" firestore:"configuration"`
	ListingType   string `json:"listingType" firestore:"listingType"` // rent or sell

	// Unit Details
	UnitNumber string `json:"unitNumber" firestore:"unitNumber"`
	Floor      string `json:"floor" firestore:"floor"`
	Location   string `json:"location" firestore:"location"`
	Address    string `json:"address" firestore:"address"`
	City       string `json:"city" firestore:"city"`
	State      string `json:"state" firestore:"state"`
	ZipCode    string `json:"zipCode" firestore:"zipCode"`

	// Area Details
	CarpetArea      string `json:"carpetArea" firestore:"carpetArea"`
	ConstructedArea string `json:"constructedArea" firestore:"constructedArea"`
	SquareFeet      int    `json:"squareFeet" firestore:"squareFeet"`

	// Tenant Information
	TenantName   string `json:"tenantName" firestore:"tenantName"`
	TenantEmail  string `json:"tenantEmail" firestore:"tenantEmail"`
	PersonName   string `json:"personName" firestore:"personName"`
	MobileNumber string `json:"mobileNumber" firestore:"mobileNumber"`
	PrimaryNo    string `json:"primaryNo" firestore:"primaryNo"`
	UltNo        string `json:"ultNo" firestore:"ultNo"`

	// Pricing Details
	MonthlyRent  string `json:"monthlyRent" firestore:"monthlyRent"`   // For rent
	SellingPrice string `json:"sellingPrice" firestore:"sellingPrice"` // For sell

	// Monthly Rent Details
	MonthlyRent1stYear string `json:"monthlyRent1stYear" firestore:"monthlyRent1stYear"`
	MonthlyRent2ndYear string `json:"monthlyRent2ndYear" firestore:"monthlyRent2ndYear"`
	MonthlyRent3rdYear string `json:"monthlyRent3rdYear" firestore:"monthlyRent3rdYear"`
	MonthlyRent4thYear string `json:"monthlyRent4thYear" firestore:"monthlyRent4thYear"`
	RentFromDate1      string `json:"rentFromDate1" firestore:"rentFromDate1"`
	RentToDate1        string `json:"rentToDate1" firestore:"rentToDate1"`
	RentFromDate2      string `json:"rentFromDate2" firestore:"rentFromDate2"`
	RentToDate2        string `json:"rentToDate2" firestore:"rentToDate2"`

	// Payment Details
	PaymentDueDate       string `json:"paymentDueDate" firestore:"paymentDueDate"`
	EscalationPercentage string `json:"escalationPercentage" firestore:"escalationPercentage"`
	EscalationAmount     string `json:"escalationAmount" firestore:"escalationAmount"`

	// Security & Agreement
	SecurityDeposit    string `json:"securityDeposit" firestore:"securityDeposit"`
	AgreementPeriod    string `json:"agreementPeriod" firestore:"agreementPeriod"`
	AgreementStartDate string `json:"agreementStartDate" firestore:"agreementStartDate"`
	AgreementEndDate   string `json:"agreementEndDate" firestore:"agreementEndDate"`

	// Notice & Lock-in
	NoticePeriod string `json:"noticePeriod" firestore:"noticePeriod"`
	LockInPeriod string `json:"lockInPeriod" firestore:"lockInPeriod"`

	// Unit Condition & Maintenance
	UnitCondition              string          `json:"unitCondition" firestore:"unitCondition"`
	MaintenanceToBePaidBy      string          `json:"maintenanceToBePaidBy" firestore:"maintenanceToBePaidBy"`
	ProjectCondition           string          `json:"projectCondition" firestore:"projectCondition"` // New Project, Ready Project, Preleased
	InternalImages             []string        `json:"internalImages" firestore:"internalImages"`
	PossessionDate             string          `json:"possessionDate" firestore:"possessionDate"`
	RentalStatus               string          `json:"rentalStatus" firestore:"rentalStatus"`       // available, rented
	AgreementStatus            string          `json:"agreementStatus" firestore:"agreementStatus"` // active, terminated, renewed, notice_served
	AnticipatedTerminationDate string          `json:"anticipatedTerminationDate" firestore:"anticipatedTerminationDate"`
	FurnishedChecklist         []FurnishedItem `json:"furnishedChecklist" firestore:"furnishedChecklist"`

	// Legacy fields for compatibility
	Description string  `json:"description" firestore:"description"`
	Price       float64 `json:"price" firestore:"price"`
	Bedrooms    int     `json:"bedrooms" firestore:"bedrooms"`
	Bathrooms   float64 `json:"bathrooms" firestore:"bathrooms"`

	// Images & Comments
	Images           []string `json:"images" firestore:"images"`
	SpecificComments string   `json:"specificComments" firestore:"specificComments"`

	// Tenants Information
	Tenants []TenantInfo `json:"tenants" firestore:"tenants"`

	// Buyers Information
	Buyers []BuyerInfo `json:"buyers" firestore:"buyers"`

	// Owner & System Info
	OwnerUID   string    `json:"ownerUID" firestore:"ownerUID"`
	OwnerName  string    `json:"ownerName" firestore:"ownerName"`
	OwnerEmail string    `json:"ownerEmail" firestore:"ownerEmail"`
	WantToSell bool      `json:"wantToSell" firestore:"wantToSell"` // Toggle for "Want to Sell?" - can be toggled ON/OFF
	Status     string    `json:"status" firestore:"status"`         // "active" or "inactive" (default: "active")
	IsActive   bool      `json:"isActive" firestore:"isActive"`     // Legacy field, kept for backward compatibility
	CreatedAt  time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type PropertyResponse struct {
	ID                    string          `json:"id"`
	Title                 string          `json:"title"`
	Description           string          `json:"description"`
	Price                 float64         `json:"price"`
	Address               string          `json:"address"`
	City                  string          `json:"city"`
	State                 string          `json:"state"`
	ZipCode               string          `json:"zipCode"`
	PropertyType          string          `json:"propertyType"`
	ListingType           string          `json:"listingType"`
	Configuration         string          `json:"configuration"`
	UnitNumber            string          `json:"unitNumber"`
	Floor                 string          `json:"floor"`
	Location              string          `json:"location"`
	CarpetArea            string          `json:"carpetArea"`
	ConstructedArea       string          `json:"constructedArea"`
	SquareFeet            int             `json:"squareFeet"`
	TenantName            string          `json:"tenantName"`
	PersonName            string          `json:"personName"`
	MobileNumber          string          `json:"mobileNumber"`
	PrimaryNo             string          `json:"primaryNo"`
	UltNo                 string          `json:"ultNo"`
	MonthlyRent           string          `json:"monthlyRent"`
	SellingPrice          string          `json:"sellingPrice"`
	MonthlyRent1stYear    string          `json:"monthlyRent1stYear"`
	MonthlyRent2ndYear    string          `json:"monthlyRent2ndYear"`
	MonthlyRent3rdYear    string          `json:"monthlyRent3rdYear"`
	MonthlyRent4thYear    string          `json:"monthlyRent4thYear"`
	RentFromDate1         string          `json:"rentFromDate1"`
	RentToDate1           string          `json:"rentToDate1"`
	RentFromDate2         string          `json:"rentFromDate2"`
	RentToDate2           string          `json:"rentToDate2"`
	PaymentDueDate        string          `json:"paymentDueDate"`
	EscalationPercentage  string          `json:"escalationPercentage"`
	EscalationAmount      string          `json:"escalationAmount"`
	SecurityDeposit       string          `json:"securityDeposit"`
	AgreementPeriod       string          `json:"agreementPeriod"`
	AgreementStartDate    string          `json:"agreementStartDate"`
	AgreementEndDate      string          `json:"agreementEndDate"`
	NoticePeriod          string          `json:"noticePeriod"`
	LockInPeriod          string          `json:"lockInPeriod"`
	UnitCondition         string          `json:"unitCondition"`
	MaintenanceToBePaidBy string          `json:"maintenanceToBePaidBy"`
	ProjectCondition      string          `json:"projectCondition"`
	InternalImages        []string        `json:"internalImages"`
	PossessionDate        string          `json:"possessionDate"`
	RentalStatus          string          `json:"rentalStatus"`
	FurnishedChecklist    []FurnishedItem `json:"furnishedChecklist"`
	Images                []string        `json:"images"`
	SpecificComments      string          `json:"specificComments"`
	Tenants               []TenantInfo    `json:"tenants"`
	Buyers                []BuyerInfo     `json:"buyers"`
	OwnerUID              string          `json:"ownerUID"`
	OwnerName             string          `json:"ownerName"`
	OwnerEmail            string          `json:"ownerEmail"`
	WantToSell            bool            `json:"wantToSell"` // Toggle for "Want to Sell?" - can be toggled ON/OFF
	Status                string          `json:"status"`     // "active" or "inactive"
	IsActive              bool            `json:"isActive"`   // Legacy field
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
	Bedrooms              int             `json:"bedrooms"`
	Bathrooms             float64         `json:"bathrooms"`
	UserRole              string          `json:"userRole,omitempty"` // "owner" or "tenant" - indicates the current user's relationship to the property
}

type CreatePropertyRequest struct {
	// Basic Property Details
	PropertyTitle string `json:"propertyTitle"`
	PropertyType  string `json:"propertyType"`
	Configuration string `json:"configuration"`
	ListingType   string `json:"listingType"`

	// Unit Details
	UnitNumber string `json:"unitNumber"`
	Floor      string `json:"floor"`
	Location   string `json:"location"`

	// Area Details
	CarpetArea      string `json:"carpetArea"`
	ConstructedArea string `json:"constructedArea"`

	// Tenant Information
	TenantName   string `json:"tenantName"`
	TenantEmail  string `json:"tenantEmail"`
	PersonName   string `json:"personName"`
	MobileNumber string `json:"mobileNumber"`
	PrimaryNo    string `json:"primaryNo"`
	UltNo        string `json:"ultNo"`

	// Pricing Details
	MonthlyRent  string `json:"monthlyRent"`
	SellingPrice string `json:"sellingPrice"`

	// Monthly Rent Details
	MonthlyRent1stYear string `json:"monthlyRent1stYear"`
	MonthlyRent2ndYear string `json:"monthlyRent2ndYear"`
	MonthlyRent3rdYear string `json:"monthlyRent3rdYear"`
	MonthlyRent4thYear string `json:"monthlyRent4thYear"`
	RentFromDate1      string `json:"rentFromDate1"`
	RentToDate1        string `json:"rentToDate1"`
	RentFromDate2      string `json:"rentFromDate2"`
	RentToDate2        string `json:"rentToDate2"`

	// Payment Details
	PaymentDueDate       string `json:"paymentDueDate"`
	EscalationPercentage string `json:"escalationPercentage"`
	EscalationAmount     string `json:"escalationAmount"`

	// Security & Agreement
	SecurityDeposit    string `json:"securityDeposit"`
	AgreementPeriod    string `json:"agreementPeriod"`
	AgreementStartDate string `json:"agreementStartDate"`
	AgreementEndDate   string `json:"agreementEndDate"`

	// Notice & Lock-in
	NoticePeriod string `json:"noticePeriod"`
	LockInPeriod string `json:"lockInPeriod"`

	// Unit Condition & Maintenance
	UnitCondition         string          `json:"unitCondition"`
	MaintenanceToBePaidBy string          `json:"maintenanceToBePaidBy"`
	ProjectCondition      string          `json:"projectCondition"`
	InternalImages        []string        `json:"internalImages"`
	PossessionDate        string          `json:"possessionDate"`
	RentalStatus          string          `json:"rentalStatus"`
	FurnishedChecklist    []FurnishedItem `json:"furnishedChecklist"`

	// Images & Comments
	Images           []string `json:"images"`
	SpecificComments string   `json:"specificComments"`

	// Owner Info
	OwnerUID string `json:"ownerUID"`
}
