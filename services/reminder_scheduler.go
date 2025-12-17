package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ReminderScheduler handles scheduling and sending visit reminders
type ReminderScheduler struct {
	emailService    *EmailService
	visitService    *VisitService
	userService     *UserService
	propertyService *PropertyService
	ticker          *time.Ticker
	rentTicker      *time.Ticker // Daily ticker for rent reminders
	stopChan        chan bool
	isRunning       bool
}

// NewReminderScheduler creates a new reminder scheduler
func NewReminderScheduler(emailService *EmailService, visitService *VisitService, userService *UserService, propertyService *PropertyService) *ReminderScheduler {
	return &ReminderScheduler{
		emailService:    emailService,
		visitService:    visitService,
		userService:     userService,
		propertyService: propertyService,
		stopChan:        make(chan bool),
		isRunning:       false,
	}
}

// Start starts the reminder scheduler that runs every minute
func (rs *ReminderScheduler) Start() {
	if rs.isRunning {
		return
	}

	rs.isRunning = true
	rs.ticker = time.NewTicker(1 * time.Minute)
	// Check for rent reminders once every 24 hours
	rs.rentTicker = time.NewTicker(24 * time.Hour)

	go func() {
		// Run immediately on start
		rs.checkAndSendReminders()
		rs.checkAndSendRentReminders()

		for {
			select {
			case <-rs.ticker.C:
				rs.checkAndSendReminders()
			case <-rs.rentTicker.C:
				rs.checkAndSendRentReminders()
			case <-rs.stopChan:
				rs.ticker.Stop()
				rs.rentTicker.Stop()
				rs.isRunning = false
				return
			}
		}
	}()
}

// Stop stops the reminder scheduler
func (rs *ReminderScheduler) Stop() {
	if rs.isRunning {
		rs.stopChan <- true
	}
}

// checkAndSendReminders checks for visits that need reminders and sends them
func (rs *ReminderScheduler) checkAndSendReminders() {

	visits, err := rs.visitService.GetPendingReminders()
	if err != nil {
		return
	}

	if len(visits) == 0 {
		return
	}

	ctx := context.Background()
	for _, visit := range visits {
		// Get user details
		user, err := rs.userService.GetUserByID(ctx, visit.UserID)
		if err != nil {
			continue
		}

		if user.Email == "" {
			continue
		}

		// Send reminder email
		err = rs.emailService.SendVisitReminder(user, &visit)
		if err != nil {
			continue
		}

		// Mark visit as notified
		err = rs.visitService.MarkAsNotified(ctx, visit.ID)
		if err != nil {
			continue
		}

	}
}

// checkAndSendRentReminders checks for active tenants who have rent due today
func (rs *ReminderScheduler) checkAndSendRentReminders() {
	ctx := context.Background()
	properties, err := rs.propertyService.GetAllProperties(ctx)
	if err != nil {
		return
	}

	today := time.Now()
	dayOfMonth := today.Day()

	for _, property := range properties {
		if !property.IsActive {
			continue
		}

		// Check each tenant
		updated := false
		for i, tenant := range property.Tenants {
			if !tenant.IsActive {
				continue
			}

			// PaymentDueDate could be "5", "05", "5th", etc.
			// innovative approach: try to extract the first number found
			dueDateStr := strings.TrimSpace(tenant.PaymentDueDate)

			// Try direct integer conversion first
			dueDate, err := strconv.Atoi(dueDateStr)
			if err != nil {
				// If error, try to extract digits from string (e.g. "5th")
				digits := ""
				for _, r := range dueDateStr {
					if r >= '0' && r <= '9' {
						digits += string(r)
					} else if len(digits) > 0 {
						// Stop after finding first sequence of digits (Assuming "5th" not "May 5")
						break
					}
				}
				if len(digits) > 0 {
					dueDate, _ = strconv.Atoi(digits)
				}
			}

			// If we successfully parsed a date (1-31)
			if dueDate > 0 && dueDate <= 31 {
				// Check if today matches
				if dayOfMonth == dueDate {
					// Check if already sent today
					lastSent := tenant.LastRentPaymentReminderSentAt
					if lastSent.IsZero() || !isSameDay(lastSent, today) {
						// Send email
						emailData := RentPaymentReminderEmailData{
							TenantName:      tenant.FirstName + " " + tenant.LastName,
							TenantEmail:     tenant.Email,
							PropertyTitle:   property.Title,
							PropertyAddress: property.Address,
							MonthlyRent:     tenant.MonthlyRent,
							DueDate:         getOrdinal(dueDate), // Use parsed date for ordinal
						}

						err := rs.emailService.SendRentPaymentReminder(emailData)
						if err == nil {
							// Update LastRentPaymentReminderSentAt
							property.Tenants[i].LastRentPaymentReminderSentAt = today
							updated = true
							fmt.Printf("Sent rent reminder to %s for property %s\n", tenant.Email, property.Title)
						} else {
							fmt.Printf("Failed to send rent reminder to %s: %v\n", tenant.Email, err)
						}
					}
				}
			}
		}

		// If any tenant was updated, save the property
		if updated {
			// We need to pass the whole tenants array to update
			// We convert struct slice to map slice or check if UpdateProperty handles struct slice
			// PropertyService.UpdateProperty takes map[string]interface{}
			// We need to convert Tenants []TenantInfo to appropriate format or simpler:
			// Modify UpdateProperty to accept updates more flexibly or convert here.
			// Let's create a map update.

			// Re-construct tenants list for update
			// Note: PropertyService.UpdateProperty handles []interface{} for tenants
			// We need to marshal or convert manually.
			// Actually simpler to just have PropertyService expose a reusable update method or do it here.
			// But PropertyService expects map.

			// Important: We modified property.Tenants in place above.
			// But to call UpdateProperty we need to pass it as map value.

			updates := map[string]interface{}{
				"tenants": property.Tenants,
			}

			_, err := rs.propertyService.UpdateProperty(ctx, property.ID, updates)
			if err != nil {
				fmt.Printf("Failed to update property %s after sending reminder: %v\n", property.ID, err)
			} else {
				fmt.Printf("Updated property %s with reminder sent status\n", property.ID)
			}
		}
	}
}

// helper check if two times are on same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// getOrdinal returns the ordinal string for a number (e.g. 1st, 2nd, 3rd, 4th)
func getOrdinal(n int) string {
	suffix := "th"
	switch n % 10 {
	case 1:
		if n%100 != 11 {
			suffix = "st"
		}
	case 2:
		if n%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if n%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

// ForceCheck forces an immediate check for reminders (useful for testing)
func (rs *ReminderScheduler) ForceCheck() {
	rs.checkAndSendReminders()
	rs.checkAndSendRentReminders()
}
