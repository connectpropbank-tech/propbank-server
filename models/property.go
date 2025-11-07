package models

import (
	"time"
)

// TenantInfo represents tenant information stored in property document
type TenantInfo struct {
	ID string `json:"id" firestore:"id"`

	// Personal Information
	FirstName        string `json:"firstName" firestore:"firstName"`
	LastName         string `json:"lastName" firestore:"lastName"`
	Email            string `json:"email" firestore:"email"`
	Phone            string `json:"phone" firestore:"phone"`
	EmergencyContact string `json:"emergencyContact" firestore:"emergencyContact"`

	// Lease Information
	LeaseStartDate  string `json:"leaseStartDate" firestore:"leaseStartDate"`
	LeaseEndDate    string `json:"leaseEndDate" firestore:"leaseEndDate"`
	MonthlyRent     string `json:"monthlyRent" firestore:"monthlyRent"`
	SecurityDeposit string `json:"securityDeposit" firestore:"securityDeposit"`

	// Background Information
	PreviousAddress  string `json:"previousAddress" firestore:"previousAddress"`
	EmploymentStatus string `json:"employmentStatus" firestore:"employmentStatus"`
	Employer         string `json:"employer" firestore:"employer"`
	MonthlyIncome    string `json:"monthlyIncome" firestore:"monthlyIncome"`

	// Additional Information
	Notes    string `json:"notes" firestore:"notes"`
	IsActive bool   `json:"isActive" firestore:"isActive"`

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
	UnitNumber   string `json:"unitNumber" firestore:"unitNumber"`
	Floor        string `json:"floor" firestore:"floor"`
	BuildingName string `json:"buildingName" firestore:"buildingName"`
	Location     string `json:"location" firestore:"location"`
	Address      string `json:"address" firestore:"address"`
	City         string `json:"city" firestore:"city"`
	State        string `json:"state" firestore:"state"`
	ZipCode      string `json:"zipCode" firestore:"zipCode"`

	// Area Details
	CarpetArea      string `json:"carpetArea" firestore:"carpetArea"`
	PlotArea        string `json:"plotArea" firestore:"plotArea"`
	ConstructedArea string `json:"constructedArea" firestore:"constructedArea"`
	SquareFeet      int    `json:"squareFeet" firestore:"squareFeet"`

	// Tenant Information
	TenantName   string `json:"tenantName" firestore:"tenantName"`
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
	UnitCondition         string `json:"unitCondition" firestore:"unitCondition"`
	MaintenanceToBePaidBy string `json:"maintenanceToBePaidBy" firestore:"maintenanceToBePaidBy"`
	ProjectCondition      string `json:"projectCondition" firestore:"projectCondition"` // New Project, Ready Project, Preleased

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
	IsActive   bool      `json:"isActive" firestore:"isActive"`
	CreatedAt  time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type PropertyResponse struct {
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	Description      string       `json:"description"`
	Price            float64      `json:"price"`
	Address          string       `json:"address"`
	City             string       `json:"city"`
	State            string       `json:"state"`
	ZipCode          string       `json:"zipCode"`
	PropertyType     string       `json:"propertyType"`
	ListingType      string       `json:"listingType"`
	ProjectCondition string       `json:"projectCondition"`
	Bedrooms         int          `json:"bedrooms"`
	Bathrooms        float64      `json:"bathrooms"`
	SquareFeet       int          `json:"squareFeet"`
	Images           []string     `json:"images"`
	Tenants          []TenantInfo `json:"tenants"`
	Buyers           []BuyerInfo  `json:"buyers"`
	OwnerUID         string       `json:"ownerUID"`
	OwnerName        string       `json:"ownerName"`
	OwnerEmail       string       `json:"ownerEmail"`
	IsActive         bool         `json:"isActive"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

type CreatePropertyRequest struct {
	// Basic Property Details
	PropertyTitle string `json:"propertyTitle"`
	PropertyType  string `json:"propertyType"`
	Configuration string `json:"configuration"`
	ListingType   string `json:"listingType"`

	// Unit Details
	UnitNumber   string `json:"unitNumber"`
	Floor        string `json:"floor"`
	BuildingName string `json:"buildingName"`
	Location     string `json:"location"`

	// Area Details
	CarpetArea      string `json:"carpetArea"`
	PlotArea        string `json:"plotArea"`
	ConstructedArea string `json:"constructedArea"`

	// Tenant Information
	TenantName   string `json:"tenantName"`
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
	UnitCondition         string `json:"unitCondition"`
	MaintenanceToBePaidBy string `json:"maintenanceToBePaidBy"`
	ProjectCondition      string `json:"projectCondition"`

	// Images & Comments
	Images           []string `json:"images"`
	SpecificComments string   `json:"specificComments"`

	// Owner Info
	OwnerUID string `json:"ownerUID"`
}
