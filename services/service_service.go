package services

import (
	"context"
	"shoprop-backend/models"
	"time"

	"cloud.google.com/go/firestore"
)

// ServiceService handles service-related operations
type ServiceService struct {
	client *firestore.Client
}

// NewServiceService creates a new ServiceService
func NewServiceService(client *firestore.Client) *ServiceService {
	return &ServiceService{
		client: client,
	}
}

// GetAllServices retrieves all active services from Firestore
func (ss *ServiceService) GetAllServices(ctx context.Context) ([]models.Service, error) {
	iter := ss.client.Collection("services").Where("isActive", "==", true).Documents(ctx)
	defer iter.Stop()

	var services []models.Service
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var service models.Service
		if err := doc.DataTo(&service); err != nil {
			continue
		}
		services = append(services, service)
	}

	return services, nil
}

// GetServiceByID retrieves a service by ID from Firestore
func (ss *ServiceService) GetServiceByID(ctx context.Context, serviceID string) (*models.Service, error) {
	doc, err := ss.client.Collection("services").Doc(serviceID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var service models.Service
	if err := doc.DataTo(&service); err != nil {
		return nil, err
	}

	return &service, nil
}

// CreateServiceRequest creates a new service request in Firestore
func (ss *ServiceService) CreateServiceRequest(ctx context.Context, serviceRequest *models.ServiceRequest) error {
	_, err := ss.client.Collection("service_requests").Doc(serviceRequest.ID).Set(ctx, serviceRequest)
	return err
}

// GetAllServiceRequests retrieves all service requests from Firestore
func (ss *ServiceService) GetAllServiceRequests(ctx context.Context) ([]models.ServiceRequest, error) {
	iter := ss.client.Collection("service_requests").OrderBy("createdAt", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var requests []models.ServiceRequest
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var request models.ServiceRequest
		if err := doc.DataTo(&request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetServiceRequestsByUser retrieves service requests for a specific user from Firestore
func (ss *ServiceService) GetServiceRequestsByUser(ctx context.Context, userUID string) ([]models.ServiceRequest, error) {
	iter := ss.client.Collection("service_requests").Where("userUID", "==", userUID).OrderBy("createdAt", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var requests []models.ServiceRequest
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var request models.ServiceRequest
		if err := doc.DataTo(&request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// UpdateServiceRequestStatus updates the status of a service request in Firestore
func (ss *ServiceService) UpdateServiceRequestStatus(ctx context.Context, requestID string, status models.ServiceStatus, adminNotes string) error {
	updates := []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updatedAt", Value: time.Now()},
	}

	if adminNotes != "" {
		updates = append(updates, firestore.Update{Path: "adminNotes", Value: adminNotes})
	}

	_, err := ss.client.Collection("service_requests").Doc(requestID).Update(ctx, updates)
	return err
}

// CreateService creates a new service in Firestore
func (ss *ServiceService) CreateService(ctx context.Context, service *models.Service) error {
	_, err := ss.client.Collection("services").Doc(service.ID).Set(ctx, service)
	return err
}

// GetUserByID retrieves a user by ID from Firestore
func (ss *ServiceService) GetUserByID(ctx context.Context, userUID string) (*models.User, error) {
	doc, err := ss.client.Collection("users").Doc(userUID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
