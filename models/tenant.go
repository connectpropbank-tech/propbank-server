package models

import "time"

// Tenant represents a tenant in the system
type Tenant struct {
	ID         string `json:"id" firestore:"id"`
	PropertyID string `json:"propertyId" firestore:"propertyId"`
	OwnerUID   string `json:"ownerUID" firestore:"ownerUID"`

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

// CreateTenantRequest represents the request payload for creating tenants
type CreateTenantRequest struct {
	PropertyID string   `json:"propertyId"`
	OwnerUID   string   `json:"ownerUID"`
	Tenants    []Tenant `json:"tenants"`
}
