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
		rs.checkAndSendNoticePeriodReminders() // Also run this on start

		for {
			select {
			case <-rs.ticker.C:
				rs.checkAndSendReminders()
			case <-rs.rentTicker.C:
				rs.checkAndSendRentReminders()
				rs.checkAndSendNoticePeriodReminders()
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
	properties, err := rs.propertyService.GetAllProperties(ctx, true)
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
					// Check if already sent today — handle both time.Time (old Timestamp records) and string (newer records)
					alreadySentToday := false
					if lastSentVal := tenant.LastRentPaymentReminderSentAt; lastSentVal != nil {
						todayStr := today.Format("2006-01-02")
						switch v := lastSentVal.(type) {
						case time.Time:
							alreadySentToday = isSameDay(v, today)
						case string:
							alreadySentToday = strings.HasPrefix(v, todayStr)
						}
					}
					if !alreadySentToday {
						// Send email
						emailData := RentPaymentReminderEmailData{
							TenantName:      tenant.FirstName + " " + tenant.LastName,
							TenantEmail:     tenant.Email,
							PropertyTitle:   property.Title,
							PropertyAddress: property.Address,
							MonthlyRent:     tenant.MonthlyRent,
							DueDate:         getOrdinal(dueDate),
						}

						err := rs.emailService.SendRentPaymentReminder(emailData)
						if err == nil {
							// Store as RFC3339 string going forward
							property.Tenants[i].LastRentPaymentReminderSentAt = today.Format(time.RFC3339)
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

// checkAndSendNoticePeriodReminders checks if notice period should start today
func (rs *ReminderScheduler) checkAndSendNoticePeriodReminders() {
	ctx := context.Background()
	properties, err := rs.propertyService.GetAllProperties(ctx, true)
	if err != nil {
		fmt.Printf("Error fetching properties for notice period check: %v\n", err)
		return
	}

	today := time.Now()

	for _, property := range properties {
		if !property.IsActive {
			continue
		}

		updated := false
		for i, tenant := range property.Tenants {
			if !tenant.IsActive {
				continue
			}

			// Ensure we have necessary dates and notice period
			if tenant.LeaseEndDate == "" || tenant.NoticePeriod == "" {
				continue
			}

			// Parse Lease End Date using resilient parser
			leaseEnd, err := parseLeaseDate(tenant.LeaseEndDate)
			if err != nil {
				// Try alternate format if needed, or log error
				// Assuming standard YYYY-MM-DD from HTML date input
				continue
			}

			// Parse Notice Period (e.g., "1 Month", "30 Days", "Manual")
			months := 0
			days := 0
			
			// Find all numbers in the string
			numStr := ""
			for _, r := range tenant.NoticePeriod {
				if r >= '0' && r <= '9' {
					numStr += string(r)
				} else if len(numStr) > 0 {
					break
				}
			}
			
			if numStr != "" {
				val, _ := strconv.Atoi(numStr)
				lowerNotice := strings.ToLower(tenant.NoticePeriod)
				if strings.Contains(lowerNotice, "day") {
					days = val
				} else {
					// Default to months if not specified or contains "month"
					months = val
				}
			}

			if months == 0 && days == 0 {
				continue
			}

			// Calculate Notice Start Date
			var noticeStartDate time.Time
			if months > 0 {
				noticeStartDate = leaseEnd.AddDate(0, -months, 0)
			} else {
				noticeStartDate = leaseEnd.AddDate(0, 0, -days)
			}

			// Check if today matches or has passed the notice start date
			if (today.After(noticeStartDate) || isSameDay(today, noticeStartDate)) && today.Before(leaseEnd) {
				// Check if already notified — handle both time.Time (old Timestamp) and string (newer records)
				alreadyNotified := false
				if lastSentVal := tenant.LastNoticePeriodReminderSentAt; lastSentVal != nil {
					switch v := lastSentVal.(type) {
					case time.Time:
						alreadyNotified = !v.IsZero()
					case string:
						alreadyNotified = v != ""
					}
				}
				if !alreadyNotified {

					// Prepare Email Data
					emailData := NoticePeriodEmailData{
						TenantName:      tenant.FirstName + " " + tenant.LastName,
						OwnerName:       property.OwnerName,
						PropertyTitle:   property.Title,
						PropertyAddress: property.Address,
						NoticePeriod:    tenant.NoticePeriod,
						LeaseEndDate:    tenant.LeaseEndDate,
					}

					// Send to Tenant
					emailData.RecipientType = "tenant"
					rs.emailService.SendNoticePeriodReminder(emailData, []string{tenant.Email})

					// Send to Owner
					emailData.RecipientType = "owner"
					if property.OwnerEmail != "" {
						rs.emailService.SendNoticePeriodReminder(emailData, []string{property.OwnerEmail})
					}

					// Store as RFC3339 string going forward
					property.Tenants[i].LastNoticePeriodReminderSentAt = today.Format(time.RFC3339)
					updated = true
					fmt.Printf("Sent notice period reminder for %s\n", property.Title)
				}
			}
		}

		if updated {
			updates := map[string]interface{}{
				"tenants": property.Tenants,
			}
			_, err := rs.propertyService.UpdateProperty(ctx, property.ID, updates)
			if err != nil {
				fmt.Printf("Failed to update property %s after sending notice reminder: %v\n", property.ID, err)
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

// ParseLeaseDate attempts to parse a date string using multiple common formats
func parseLeaseDate(dateStr string) (time.Time, error) {
	formats := []string{"2006-01-02", "02/01/2006", "02-01-2006", "2006/01/02"}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// ForceCheck forces an immediate check for reminders (useful for testing)
func (rs *ReminderScheduler) ForceCheck() {
	fmt.Println("[ReminderScheduler] Force triggering all reminder checks...")
	rs.checkAndSendReminders()
	rs.checkAndSendRentReminders()
	rs.checkAndSendNoticePeriodReminders()
}
