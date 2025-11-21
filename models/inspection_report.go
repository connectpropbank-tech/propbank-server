package models

import (
	"time"
)

// InspectionReport represents an inspection report in the database
type InspectionReport struct {
	ID            string `firestore:"id" json:"id"`
	PropertyID    string `firestore:"propertyId" json:"propertyId"`
	PropertyTitle string `firestore:"propertyTitle" json:"propertyTitle"`
	UserID        string `firestore:"userId" json:"userId"`
	UserName      string `firestore:"userName" json:"userName"`
	UserEmail     string `firestore:"userEmail" json:"userEmail"`
	UserPhone     string `firestore:"userPhone" json:"userPhone"`
	ReportType    string `firestore:"reportType" json:"reportType"` // "on_possession" or "on_handover"
	Report        string `firestore:"report" json:"report"`         // The actual report text

	// On Possession / On Handover Details
	PossessionLetterFile string `firestore:"possessionLetterFile,omitempty" json:"possessionLetterFile,omitempty"` // Uploaded file
	HandoverLetterFile   string `firestore:"handoverLetterFile,omitempty" json:"handoverLetterFile,omitempty"`     // Uploaded file
	KeysDetails          string `firestore:"keysDetails,omitempty" json:"keysDetails,omitempty"`

	// Document attachments
	ElectricityBillMeterImage string `firestore:"electricityBillMeterImage,omitempty" json:"electricityBillMeterImage,omitempty"`
	ElectricityBillReceipt    string `firestore:"electricityBillReceipt,omitempty" json:"electricityBillReceipt,omitempty"`
	ApartmentConditionImage   string `firestore:"apartmentConditionImage,omitempty" json:"apartmentConditionImage,omitempty"`
	MGLBillMeterImage         string `firestore:"mglBillMeterImage,omitempty" json:"mglBillMeterImage,omitempty"`
	MGLBillReceipt            string `firestore:"mglBillReceipt,omitempty" json:"mglBillReceipt,omitempty"`
	InternetImage             string `firestore:"internetImage,omitempty" json:"internetImage,omitempty"`
	InternetReceipt           string `firestore:"internetReceipt,omitempty" json:"internetReceipt,omitempty"`
	OtherDetails              string `firestore:"otherDetails,omitempty" json:"otherDetails,omitempty"` // For "If any other, add option"

	CreatedAt time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updatedAt" json:"updatedAt"`
}

// CreateInspectionReportRequest represents the request body for creating an inspection report
type CreateInspectionReportRequest struct {
	PropertyID string `json:"propertyId" validate:"required"`
	ReportType string `json:"reportType" validate:"required,oneof=on_possession on_handover"`
	Report     string `json:"report" validate:"required"`

	// On Possession / On Handover Details
	PossessionLetterFile string `json:"possessionLetterFile,omitempty"`
	HandoverLetterFile   string `json:"handoverLetterFile,omitempty"`
	KeysDetails          string `json:"keysDetails,omitempty"`

	// Document attachments
	ElectricityBillMeterImage string `json:"electricityBillMeterImage,omitempty"`
	ElectricityBillReceipt    string `json:"electricityBillReceipt,omitempty"`
	ApartmentConditionImage   string `json:"apartmentConditionImage,omitempty"`
	MGLBillMeterImage         string `json:"mglBillMeterImage,omitempty"`
	MGLBillReceipt            string `json:"mglBillReceipt,omitempty"`
	InternetImage             string `json:"internetImage,omitempty"`
	InternetReceipt           string `json:"internetReceipt,omitempty"`
	OtherDetails              string `json:"otherDetails,omitempty"`
}
