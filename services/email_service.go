package services

import (
	"context"
	"fmt"
	"os"
	"shoprop-backend/models"
	"strings"
	"time"

	"github.com/resend/resend-go/v2"
)

// India Standard Time (IST) is UTC+5:30
var IST = time.FixedZone("IST", 5*60*60+30*60)

type EmailService struct {
	resendClient *resend.Client
	FromEmail    string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	client := resend.NewClient(apiKey)

	return &EmailService{
		resendClient: client,
		FromEmail:    getEnvOrDefault("FROM_EMAIL", "connect@propbank.shop"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// FormatTimeIST formats a time in IST timezone
func FormatTimeIST(t time.Time) string {
	return t.In(IST).Format("Monday, January 2, 2006 at 3:04 PM IST")
}

// GetCurrentTimeIST returns current time in IST
func GetCurrentTimeIST() time.Time {
	return time.Now().In(IST)
}

// SendVisitReminder sends an email reminder for a scheduled visit
func (es *EmailService) SendVisitReminder(user *models.User, visit *models.Visit) error {
	subject := fmt.Sprintf("🏠 Reminder: %s", visit.Title)

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <!-- Header -->
        <tr>
            <td style="background: linear-gradient(135deg, #1a365d 0%%, #2563eb 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #e0e7ff; margin: 10px 0 0 0; font-size: 14px;">Property Visit Reminder</p>
            </td>
        </tr>
        
        <!-- Main Content -->
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    This is a friendly reminder about your upcoming property visit:
                </p>
                
                <!-- Visit Details Card -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0;">
                    <tr>
                        <td style="padding: 25px;">
                            <table cellpadding="0" cellspacing="0" width="100%%">
                                <tr>
                                    <td style="padding: 10px 0; border-bottom: 1px solid #e2e8f0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">Visit Title</strong>
                                        <p style="color: #1e293b; font-size: 18px; margin: 5px 0 0 0; font-weight: 600;">%s</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 15px 0; border-bottom: 1px solid #e2e8f0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">📅 Date & Time</strong>
                                        <p style="color: #1e293b; font-size: 16px; margin: 5px 0 0 0;">%s</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 15px 0 0 0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">📝 Details</strong>
                                        <p style="color: #1e293b; font-size: 14px; margin: 5px 0 0 0; line-height: 1.5;">%s</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>
                
                <div style="margin-top: 25px; padding: 20px; background-color: #fef3c7; border-radius: 8px; border-left: 4px solid #f59e0b;">
                    <p style="color: #92400e; font-size: 14px; margin: 0; font-weight: 500;">
                        ⏰ Please make sure to arrive on time for your scheduled visit.
                    </p>
                </div>
                
                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 25px 0 0 0;">
                    If you need to reschedule or cancel, please update your visit in the Propbank app.
                </p>
            </td>
        </tr>
        
        <!-- Footer -->
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. If you have any questions, contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">
                    © 2025 Propbank. All rights reserved.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		user.Name,
		visit.Title,
		FormatTimeIST(visit.VisitDate),
		visit.Description,
	)

	return es.sendEmail(user.Email, subject, body)
}

// SendVisitConfirmation sends a confirmation email when a visit is scheduled
func (es *EmailService) SendVisitConfirmation(user *models.User, visit *models.Visit) error {
	subject := fmt.Sprintf("✅ Visit Scheduled: %s", visit.Title)

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <!-- Header -->
        <tr>
            <td style="background: linear-gradient(135deg, #059669 0%%, #10b981 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 14px;">Visit Confirmation</p>
            </td>
        </tr>
        
        <!-- Main Content -->
        <tr>
            <td style="padding: 40px 30px;">
                <div style="text-align: center; margin-bottom: 30px;">
                    <div style="width: 60px; height: 60px; background-color: #d1fae5; border-radius: 50%%; display: inline-flex; align-items: center; justify-content: center;">
                        <span style="font-size: 30px;">✓</span>
                    </div>
                </div>
                
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px; text-align: center;">Visit Scheduled Successfully!</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0; text-align: center;">
                    Hi %s, your property visit has been confirmed.
                </p>
                
                <!-- Visit Details Card -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0;">
                    <tr>
                        <td style="padding: 25px;">
                            <table cellpadding="0" cellspacing="0" width="100%%">
                                <tr>
                                    <td style="padding: 10px 0; border-bottom: 1px solid #e2e8f0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">Visit Title</strong>
                                        <p style="color: #1e293b; font-size: 18px; margin: 5px 0 0 0; font-weight: 600;">%s</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 15px 0; border-bottom: 1px solid #e2e8f0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">📅 Date & Time</strong>
                                        <p style="color: #1e293b; font-size: 16px; margin: 5px 0 0 0;">%s</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 15px 0; border-bottom: 1px solid #e2e8f0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">🔔 Reminder</strong>
                                        <p style="color: #1e293b; font-size: 14px; margin: 5px 0 0 0;">%d minutes before via %s</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 15px 0 0 0;">
                                        <strong style="color: #64748b; font-size: 12px; text-transform: uppercase;">📝 Details</strong>
                                        <p style="color: #1e293b; font-size: 14px; margin: 5px 0 0 0; line-height: 1.5;">%s</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>
                
                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 25px 0 0 0;">
                    You can manage your visits anytime by logging into the Propbank app.
                </p>
            </td>
        </tr>
        
        <!-- Footer -->
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. If you have any questions, contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">
                    © 2025 Propbank. All rights reserved.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		user.Name,
		visit.Title,
		FormatTimeIST(visit.VisitDate),
		visit.ReminderTime,
		visit.ReminderType,
		visit.Description,
	)

	return es.sendEmail(user.Email, subject, body)
}

// SendGeneralInquiryNotification sends an email to admin for a new quick inquiry
func (es *EmailService) SendGeneralInquiryNotification(adminEmail string, req models.AdminNotification) error {
	subject := fmt.Sprintf("📩 New Quick Inquiry: %s", req.UserName)

	listingTypeDisplay := "Renting"
	if req.InquiryType == "buy" {
		listingTypeDisplay = "Buying"
	} else if req.InquiryType == "sell" {
		listingTypeDisplay = "Selling"
	}

	visitInfo := "Not requested"
	if req.RequestVisit {
		visitInfo = fmt.Sprintf("Requested on %s at %s", req.VisitDate, req.VisitTime)
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f8fafc;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 20px auto; background-color: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);">
        <!-- Header -->
        <tr>
            <td style="background: linear-gradient(135deg, #1e293b 0%%, #334155 100%%); padding: 40px 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 24px; font-weight: 700; letter-spacing: -0.025em;">🏠 Propbank</h1>
                <p style="color: #94a3b8; margin: 10px 0 0 0; font-size: 14px; font-weight: 500;">New Quick Inquiry Received</p>
            </td>
        </tr>
        
        <!-- Main Content -->
        <tr>
            <td style="padding: 40px 30px;">
                <div style="margin-bottom: 30px;">
                    <h2 style="color: #0f172a; margin: 0 0 10px 0; font-size: 20px; font-weight: 600;">Full Inquiry Details</h2>
                    <p style="color: #64748b; font-size: 15px; line-height: 1.5; margin: 0;">A potential client has just submitted an inquiry through the Quick Inquiry form.</p>
                </div>
                
                <!-- Client Info Card -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f1f5f9; border-radius: 12px; margin-bottom: 24px;">
                    <tr>
                        <td style="padding: 24px;">
                            <h3 style="color: #475569; margin: 0 0 16px 0; font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em;">👤 Client Information</h3>
                            <table cellpadding="0" cellspacing="0" width="100%%">
                                <tr>
                                    <td style="padding: 4px 0; color: #64748b; font-size: 14px; width: 100px;">Name:</td>
                                    <td style="padding: 4px 0; color: #0f172a; font-size: 14px; font-weight: 600;">%s</td>
                                </tr>
                                <tr>
                                    <td style="padding: 4px 0; color: #64748b; font-size: 14px;">Email:</td>
                                    <td style="padding: 4px 0; color: #0f172a; font-size: 14px; font-weight: 500;">%s</td>
                                </tr>
                                <tr>
                                    <td style="padding: 4px 0; color: #64748b; font-size: 14px;">Phone:</td>
                                    <td style="padding: 4px 0; color: #0f172a; font-size: 14px; font-weight: 500;">%s</td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>

                <!-- Inquiry Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f1f5f9; border-radius: 12px; margin-bottom: 24px;">
                    <tr>
                        <td style="padding: 24px;">
                            <h3 style="color: #475569; margin: 0 0 16px 0; font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em;">🏢 Property Interest</h3>
                            <table cellpadding="0" cellspacing="0" width="100%%">
                                <tr>
                                    <td style="padding: 4px 0; color: #64748b; font-size: 14px; width: 100px;">Interested in:</td>
                                    <td style="padding: 4px 0;"><span style="background-color: #cbd5e1; color: #0f172a; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 700;">%s</span></td>
                                </tr>
                                <tr>
                                    <td style="padding: 4px 0; color: #64748b; font-size: 14px;">Type:</td>
                                    <td style="padding: 4px 0; color: #0f172a; font-size: 14px; font-weight: 600; text-transform: capitalize;">%s</td>
                                </tr>
                                <tr>
                                    <td style="padding: 4px 0; color: #64748b; font-size: 14px;">Visit:</td>
                                    <td style="padding: 4px 0; color: #0f172a; font-size: 14px; font-weight: 500;">%s</td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>

                <div style="background-color: #fff7ed; border-left: 4px solid #f97316; padding: 20px; border-radius: 4px;">
                    <p style="color: #7c2d12; font-size: 14px; margin: 0; line-height: 1.6;">
                        <strong>Message:</strong><br/>
                        %s
                    </p>
                </div>
                
                <div style="margin-top: 40px; text-align: center;">
                    <a href="mailto:%s" style="display: inline-block; background-color: #2563eb; color: #ffffff; padding: 12px 30px; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 15px;">Reply to Client</a>
                </div>
            </td>
        </tr>
        
        <!-- Footer -->
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #94a3b8; font-size: 12px; margin: 0;">Sent automatically by Propbank Admin System.</p>
                <p style="color: #cbd5e1; font-size: 11px; margin: 8px 0 0 0;">© 2026 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		req.UserName,
		req.UserEmail,
		req.UserPhone,
		listingTypeDisplay,
		req.PropertyType,
		visitInfo,
		req.Message,
		req.UserEmail,
	)

	return es.sendEmail(adminEmail, subject, body)
}

// sendEmail sends an email using Resend API
func (es *EmailService) sendEmail(to, subject, body string) error {
	fmt.Printf("[EmailService] Attempting to send email to %s, Subject: %s\n", to, subject)

	// If API key is not set, skip sending (useful for dev/test)
	if os.Getenv("RESEND_API_KEY") == "" {
		fmt.Printf("[EmailService] RESEND_API_KEY not set, skipping email to %s\n", to)
		return nil
	}

	// Ensure all hardcoded instances of the legacy email are replaced with the correct one
	body = strings.ReplaceAll(body, "connectpropbank@gmail.com", es.FromEmail)

	fmt.Printf("[EmailService] Sending email to %s using sender %s\n", to, es.FromEmail)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Propbank <%s>", es.FromEmail),
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := es.resendClient.Emails.Send(params)
	if err != nil {
		fmt.Printf("[EmailService] ERROR sending to %s: %v\n", to, err)
		return fmt.Errorf("failed to send email via Resend: %v", err)
	}

	fmt.Printf("[EmailService] SUCCESS: Email sent to %s\n", to)
	return nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CheckAndSendReminders checks for visits that need reminders and sends them
func (es *EmailService) CheckAndSendReminders(visitService *VisitService, userService *UserService) {

	// This would typically be called by a cron job or background service
	// For now, it's a placeholder for the reminder system
	visits, err := visitService.GetPendingReminders()
	if err != nil {
		return
	}

	ctx := context.Background()
	for _, visit := range visits {
		user, err := userService.GetUserByID(ctx, visit.UserID)
		if err != nil {
			continue
		}

		// Check if it's time to send reminder
		reminderTime := visit.VisitDate.Add(-time.Duration(visit.ReminderTime) * time.Minute)
		if time.Now().After(reminderTime) && visit.NotifiedAt.IsZero() {
			if err := es.SendVisitReminder(user, &visit); err == nil {
				// Mark as notified
				visit.NotifiedAt = time.Now()
				visitService.UpdateVisit(ctx, visit.ID, &models.UpdateVisitRequest{})
			}
		}
	}
}

// TenantAddedEmailData contains data for tenant added email
type TenantAddedEmailData struct {
	TenantName      string
	TenantEmail     string
	TenantPhone     string
	OwnerName       string
	OwnerEmail      string
	OwnerPhone      string
	PropertyTitle   string
	PropertyAddress string
	PropertyType    string
	MonthlyRent     string
	AgreementStart  string
	AgreementEnd    string
}

// SendTenantAddedNotification sends email to both owner and tenant when tenant is added
func (es *EmailService) SendTenantAddedNotification(data TenantAddedEmailData) error {
	fmt.Printf("[EmailService] Starting SendTenantAddedNotification for Tenant: %s (%s) and Owner: %s (%s)\n", data.TenantName, data.TenantEmail, data.OwnerName, data.OwnerEmail)
	var errs []error

	// Send email to Tenant
	tenantSubject := "🏠 Welcome! You've been added as a Tenant - Propbank"
	tenantBody := es.buildTenantAddedEmail(data, true)

	if data.TenantEmail != "" {
		if err := es.sendEmail(data.TenantEmail, tenantSubject, tenantBody); err != nil {
			fmt.Printf("[EmailService] Failed to send tenant notification: %v\n", err)
			errs = append(errs, fmt.Errorf("tenant email failed: %v", err))
		}
	}

	// Send email to Owner
	ownerSubject := "🏠 Tenant Added to Your Property - Propbank"
	ownerBody := es.buildTenantAddedEmail(data, false)

	if data.OwnerEmail != "" {
		if err := es.sendEmail(data.OwnerEmail, ownerSubject, ownerBody); err != nil {
			fmt.Printf("[EmailService] Failed to send owner notification: %v\n", err)
			errs = append(errs, fmt.Errorf("owner email failed: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("email notification delivery failed: %v", errs)
	}

	return nil
}

// buildTenantAddedEmail builds the HTML email for tenant added notification
func (es *EmailService) buildTenantAddedEmail(data TenantAddedEmailData, isForTenant bool) string {
	recipientName := data.OwnerName
	headerText := "Tenant Added to Your Property"
	introText := "A new tenant has been added to your property."

	if isForTenant {
		recipientName = data.TenantName
		headerText = "Welcome to Your New Home!"
		introText = "You have been added as a tenant to the following property."
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #059669 0%%, #10b981 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 14px;">%s</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">%s</p>
                
                <!-- Property Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1a365d; margin: 0 0 15px 0; font-size: 16px;">🏢 Property Details</h3>
                            <p style="margin: 5px 0;"><strong>Property:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Address:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Type:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Monthly Rent:</strong> ₹%s</p>
                        </td>
                    </tr>
                </table>



                <!-- Agreement Period -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f0fdf4; border-radius: 12px; border: 1px solid #86efac;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #166534; margin: 0 0 15px 0; font-size: 16px;">📅 Agreement Period</h3>
                            <p style="margin: 5px 0;"><strong>Start Date:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>End Date:</strong> %s</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. Contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		headerText,
		recipientName,
		introText,
		data.PropertyTitle,
		data.PropertyAddress,
		data.PropertyType,
		data.MonthlyRent,

		data.AgreementStart,
		data.AgreementEnd,
	)
}

// AgreementTerminationEmailData contains data for agreement termination email
type AgreementTerminationEmailData struct {
	TenantName         string
	TenantEmail        string
	TenantPhone        string
	OwnerName          string
	OwnerEmail         string
	OwnerPhone         string
	PropertyTitle      string
	PropertyAddress    string
	PropertyType       string
	MonthlyRent        string // Added missing field
	TerminationDate    string
	Reason             string
	AgreementStartDate string
	AgreementEndDate   string
	AgreementPeriod    string
	RaisedBy           string // Added RaisedBy field
}

// SendAgreementTerminationNotification sends email to both owner and tenant when agreement is terminated
func (es *EmailService) SendAgreementTerminationNotification(data AgreementTerminationEmailData) error {
	// Send email to Tenant
	tenantSubject := "⚠️ Agreement Termination Notice - Propbank"
	tenantBody := es.buildAgreementTerminationEmail(data, true)

	if data.TenantEmail != "" {
		if err := es.sendEmail(data.TenantEmail, tenantSubject, tenantBody); err != nil {
		} else {
		}
	}

	// Send email to Owner
	ownerSubject := "✅ Agreement Terminated Successfully - Propbank"
	ownerBody := es.buildAgreementTerminationEmail(data, false)

	if data.OwnerEmail != "" {
		if err := es.sendEmail(data.OwnerEmail, ownerSubject, ownerBody); err != nil {
			return err
		}
	}

	return nil
}

// SendAgreementNoticeNotification sends email to both owner and tenant when notice is served
func (es *EmailService) SendAgreementNoticeNotification(data AgreementTerminationEmailData) error {
	// Send email to Tenant
	tenantSubject := "📅 Agreement Termination Notice Served - Propbank"
	tenantBody := es.buildAgreementNoticeEmail(data, true)

	if data.TenantEmail != "" {
		if err := es.sendEmail(data.TenantEmail, tenantSubject, tenantBody); err != nil {
		}
	}

	// Send email to Owner
	ownerSubject := "✅ Notice Period Initiated - Propbank"
	ownerBody := es.buildAgreementNoticeEmail(data, false)

	if data.OwnerEmail != "" {
		if err := es.sendEmail(data.OwnerEmail, ownerSubject, ownerBody); err != nil {
		}
	}

	return nil
}

func (es *EmailService) buildAgreementNoticeEmail(data AgreementTerminationEmailData, isTenant bool) string {
	var headerText, recipientName, introText string

	if isTenant {
		headerText = "Termination Notice"
		recipientName = data.TenantName
		introText = fmt.Sprintf("We have received a notice to terminate the lease agreement for <strong>%s</strong>. The agreement will be terminated on <strong>%s</strong>.", data.PropertyTitle, data.TerminationDate)
	} else {
		headerText = "Notice Initiated"
		recipientName = data.OwnerName
		introText = fmt.Sprintf("You have successfully served a termination notice for <strong>%s</strong>. The agreement is scheduled to end on <strong>%s</strong>.", data.PropertyTitle, data.TerminationDate)
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <!-- Header -->
        <tr>
            <td style="background: linear-gradient(135deg, #f59e0b 0%%, #d97706 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #fde68a; margin: 10px 0 0 0; font-size: 14px;">%s</p>
            </td>
        </tr>
        
        <!-- Main Content -->
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    %s
                </p>
                
                <!-- Property Details Card -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #fffbeb; border-radius: 12px; border: 1px solid #fcd34d;">
                    <tr>
                        <td style="padding: 25px;">
                            <h3 style="color: #92400e; margin: 0 0 15px 0; font-size: 16px; border-bottom: 1px solid #fcd34d; padding-bottom: 10px;">
                                🏢 Property Details
                            </h3>
                            <table cellpadding="0" cellspacing="0" width="100%%">
                                <tr>
                                    <td style="padding: 5px 0;"><strong style="color: #78350f;">Title:</strong></td>
                                    <td style="color: #1e293b;">%s</td>
                                </tr>
                                <tr>
                                    <td style="padding: 5px 0;"><strong style="color: #78350f;">Address:</strong></td>
                                    <td style="color: #1e293b;">%s</td>
                                </tr>
                                <tr>
                                    <td style="padding: 5px 0;"><strong style="color: #78350f;">Type:</strong></td>
                                    <td style="color: #1e293b;">%s</td>
                                </tr>
                                <tr>
                                    <td style="padding: 5px 0;"><strong style="color: #78350f;">Rent:</strong></td>
                                    <td style="color: #1e293b;">₹%s/month</td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>

                <!-- Schedule Card -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="margin-top: 20px; background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0;">
                    <tr>
                        <td style="padding: 25px;">
                            <h3 style="color: #475569; margin: 0 0 15px 0; font-size: 16px; border-bottom: 1px solid #e2e8f0; padding-bottom: 10px;">
                                📅 Termination Schedule
                            </h3>
                            <p style="color: #1e293b; font-size: 18px; margin: 0; font-weight: 600; text-align: center;">
                                %s
                            </p>
                             <p style="color: #64748b; font-size: 13px; margin: 10px 0 0 0; text-align: center;">
                                Anticipated Termination Date
                            </p>
                            <p style="color: #64748b; font-size: 14px; margin: 15px 0 0 0; text-align: center;">
                                <strong>Notice Raised By:</strong> %s
                            </p>
                        </td>
                    </tr>
                </table>
                
                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 25px 0 0 0;">
                    Please ensure all dues are cleared and the property is vacated by the termination date.
                </p>
            </td>
        </tr>
        
        <!-- Footer -->
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. If you have any questions, contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		headerText,
		recipientName,
		introText,
		data.PropertyTitle,
		data.PropertyAddress,
		data.PropertyType,
		data.MonthlyRent,
		data.TerminationDate,
		data.RaisedBy, // Added RaisedBy
	)
}

// buildAgreementTerminationEmail builds the HTML email for agreement termination notification
func (es *EmailService) buildAgreementTerminationEmail(data AgreementTerminationEmailData, isForTenant bool) string {
	recipientName := data.OwnerName
	headerText := "Agreement Terminated"
	headerColor := "#059669"
	introText := "The rental agreement has been successfully terminated."

	if isForTenant {
		recipientName = data.TenantName
		headerText = "Agreement Termination Notice"
		headerColor = "#dc2626"
		introText = "The property owner has terminated your rental agreement. Please review the details below."
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, %s 0%%, #ef4444 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #fee2e2; margin: 10px 0 0 0; font-size: 14px;">%s</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">%s</p>
                
                <!-- Termination Notice -->
                <div style="background-color: #fef2f2; border: 1px solid #fecaca; border-radius: 12px; padding: 20px; margin-bottom: 20px;">
                    <h3 style="color: #dc2626; margin: 0 0 10px 0; font-size: 16px;">⚠️ Termination Details</h3>
                    <p style="margin: 5px 0;"><strong>Termination Date:</strong> %s</p>
                    <p style="margin: 5px 0;"><strong>Reason:</strong> %s</p>
                    <p style="margin: 5px 0;"><strong>Notice Raised By:</strong> %s</p>
                </div>

                <!-- Property Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1a365d; margin: 0 0 15px 0; font-size: 16px;">🏢 Property Details</h3>
                            <p style="margin: 5px 0;"><strong>Property:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Address:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Type:</strong> %s</p>
                        </td>
                    </tr>
                </table>



                <!-- Agreement Period -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #ecfdf5; border-radius: 12px; border: 1px solid #6ee7b7;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #065f46; margin: 0 0 15px 0; font-size: 16px;">📅 Agreement Period</h3>
                            <p style="margin: 5px 0;"><strong>Start Date:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>End Date:</strong> %s</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. Contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		headerColor,
		headerText,
		recipientName,
		introText,
		data.TerminationDate,
		data.Reason,
		data.RaisedBy,
		data.PropertyTitle,
		data.PropertyAddress,
		data.PropertyType,

		data.AgreementStartDate,
		data.AgreementEndDate,
	)
}

// RentPaymentReminderEmailData contains data for rent payment reminder email
type RentPaymentReminderEmailData struct {
	TenantName      string
	TenantEmail     string
	PropertyTitle   string
	PropertyAddress string
	MonthlyRent     string
	DueDate         string
	PaymentLink     string // Optional
}

// SendRentPaymentReminder sends email to tenant for rent payment reminder
func (es *EmailService) SendRentPaymentReminder(data RentPaymentReminderEmailData) error {
	subject := "🏠 Rent Payment Due Reminder - Propbank"
	body := es.buildRentPaymentReminderEmail(data)

	if data.TenantEmail != "" {
		return es.sendEmail(data.TenantEmail, subject, body)
	}
	return nil
}

// buildRentPaymentReminderEmail builds the HTML email for rent payment reminder
func (es *EmailService) buildRentPaymentReminderEmail(data RentPaymentReminderEmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #4f46e5 0%%, #818cf8 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #e0e7ff; margin: 10px 0 0 0; font-size: 14px;">Rent Payment Reminder</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    This is a friendly reminder that your rent payment for the following property is due today.
                </p>
                
                <!-- Payment Details -->
                <div style="background-color: #f0fdf4; border: 1px solid #86efac; border-radius: 12px; padding: 25px; margin-bottom: 25px; text-align: center;">
                    <p style="color: #166534; font-size: 14px; margin: 0 0 5px 0; text-transform: uppercase; font-weight: 600;">Amount Due</p>
                    <h1 style="color: #15803d; margin: 0; font-size: 36px;">₹%s</h1>
                    <p style="color: #166534; font-size: 14px; margin: 5px 0 0 0;">Due Date: %s of every month</p>
                </div>

                <!-- Property Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1a365d; margin: 0 0 15px 0; font-size: 16px;">🏢 Property Details</h3>
                            <p style="margin: 5px 0;"><strong>Property:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Address:</strong> %s</p>
                        </td>
                    </tr>
                </table>
                
                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 25px 0 0 0;">
                    Please ensure the payment is made to the property owner to avoid any late fees.
                </p>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. Contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		data.TenantName,
		data.MonthlyRent,
		data.DueDate,
		data.PropertyTitle,
		data.PropertyAddress,
	)
}

type NoticePeriodEmailData struct {
	TenantName      string
	OwnerName       string
	PropertyTitle   string
	PropertyAddress string
	NoticePeriod    string
	LeaseEndDate    string
	RecipientType   string // "tenant" or "owner"
}

func (es *EmailService) SendNoticePeriodReminder(data NoticePeriodEmailData, recipients []string) error {
	var subject string
	var body string

	if data.RecipientType == "tenant" {
		subject = fmt.Sprintf("Notice Period Started for %s", data.PropertyTitle)
		body = es.buildTenantNoticePeriodEmail(data)
	} else {
		subject = fmt.Sprintf("Notice Period Reminder for %s", data.PropertyTitle)
		body = es.buildOwnerNoticePeriodEmail(data)
	}

	for _, to := range recipients {
		if err := es.sendEmail(to, subject, body); err != nil {
			return err
		}
	}
	return nil
}

func (es *EmailService) buildTenantNoticePeriodEmail(data NoticePeriodEmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #f59e0b 0%%, #fbbf24 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #fffbeb; margin: 10px 0 0 0; font-size: 14px;">Notice Period Activation</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    This email is to inform you that the <strong>%s</strong> notice period for your lease has officially started.
                </p>
                 <div style="background-color: #fffbeb; border: 1px solid #fcd34d; border-radius: 12px; padding: 25px; margin-bottom: 25px; text-align: center;">
                    <p style="color: #92400e; font-size: 14px; margin: 0 0 5px 0; text-transform: uppercase; font-weight: 600;">Lease End Date</p>
                    <h1 style="color: #b45309; margin: 0; font-size: 28px;">%s</h1>
                    <p style="color: #92400e; font-size: 14px; margin: 5px 0 0 0;">Notice Period: %s</p>
                </div>

                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 25px 0 0 0;">
                    Please coordinate with the property owner regarding move-out procedures or lease renewal discussions if applicable.
                </p>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`, data.TenantName, data.NoticePeriod, data.LeaseEndDate, data.NoticePeriod)
}

func (es *EmailService) buildOwnerNoticePeriodEmail(data NoticePeriodEmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
             <td style="background: linear-gradient(135deg, #3b82f6 0%%, #60a5fa 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #bfdbfe; margin: 10px 0 0 0; font-size: 14px;">Tenant Notice Period Alert</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    The <strong>%s</strong> notice period for your property has started.
                </p>
                
                 <div style="background-color: #eff6ff; border: 1px solid #bfdbfe; border-radius: 12px; padding: 20px; margin-bottom: 25px;">
                    <p style="margin: 5px 0; color: #1e3a8a;"><strong>Tenant:</strong> %s</p>
                    <p style="margin: 5px 0; color: #1e3a8a;"><strong>Property:</strong> %s</p>
                    <p style="margin: 5px 0; color: #1e3a8a;"><strong>Lease End Date:</strong> %s</p>
                </div>

                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 25px 0 0 0;">
                    It is recommended to start looking for new tenants or discuss renewal options with the current tenant.
                </p>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`, data.OwnerName, data.NoticePeriod, data.TenantName, data.PropertyTitle, data.LeaseEndDate)
}

// SendTerminationRequestNotification sends email to owner when tenant requests termination
func (es *EmailService) SendTerminationRequestNotification(ownerEmail, ownerName, tenantName, tenantEmail, propertyTitle, propertyID, noticePeriod string) error {
	if noticePeriod == "" {
		noticePeriod = "Immediate"
	}

	// 1. Send Email to Owner
	ownerSubject := fmt.Sprintf("📢 Termination Request: %s - %s", propertyTitle, noticePeriod)
	ownerBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #ef4444 0%%, #dc2626 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #fecaca; margin: 10px 0 0 0; font-size: 14px;">Termination Request</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    Your tenant, <strong>%s</strong>, has requested to terminate the lease agreement for <strong>%s</strong>.
                </p>
                
                <div style="background-color: #fef2f2; border: 1px solid #fee2e2; border-radius: 8px; padding: 20px; margin-bottom: 25px;">
                    <p style="color: #991b1b; font-size: 14px; margin: 0; font-weight: 500;">
                        Requested Notice Period: <strong>%s</strong>
                    </p>
                    <p style="color: #991b1b; font-size: 14px; margin: 10px 0 0 0; font-weight: 500;">
                        Action Required: Please review this request and take appropriate action (Serve Notice or Terminate) through the Propbank dashboard.
                    </p>
                </div>

                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 0;">
                    Property ID: %s
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, ownerName, tenantName, propertyTitle, noticePeriod, propertyID)

	if err := es.sendEmail(ownerEmail, ownerSubject, ownerBody); err != nil {
		// Log but continue
	}

	// 2. Send Confirmation Email to Tenant
	if tenantEmail != "" {
		tenantSubject := "Termination Request Received - Propbank"
		tenantBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #3b82f6 0%%, #2563eb 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #bfdbfe; margin: 10px 0 0 0; font-size: 14px;">Termination Request Sent</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    We have received your request to terminate the agreement for <strong>%s</strong>.
                </p>
                
                 <div style="background-color: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px; padding: 20px; margin-bottom: 25px;">
                    <p style="color: #1e40af; font-size: 14px; margin: 0; font-weight: 500;">
                        Requested Notice Period: <strong>%s</strong>
                    </p>
                     <p style="color: #1e40af; font-size: 14px; margin: 10px 0 0 0;">
                        The owner (%s) has been notified. They will process your request shortly.
                    </p>
                </div>
            </td>
        </tr>
    </table>
</body>
</html>
`, tenantName, propertyTitle, noticePeriod, ownerName)

		if err := es.sendEmail(tenantEmail, tenantSubject, tenantBody); err != nil {
			// Log but continue
		}
	}

	return nil
}

// PaymentDueUpdateEmailData contains data for payment due date update email
type PaymentDueUpdateEmailData struct {
	TenantName      string
	TenantEmail     string
	PropertyTitle   string
	PropertyAddress string
	NewDueDate      string
	OwnerName       string
	OwnerEmail      string
	OwnerPhone      string
}

// SendPaymentDueUpdateNotification sends email to tenant when payment due date is updated
func (es *EmailService) SendPaymentDueUpdateNotification(data PaymentDueUpdateEmailData) error {
	subject := "📅 Payment Due Date Updated - Propbank"
	body := es.buildPaymentDueUpdateEmail(data)

	if data.TenantEmail != "" {
		if err := es.sendEmail(data.TenantEmail, subject, body); err != nil {
			return err
		}
	}
	return nil
}

// buildPaymentDueUpdateEmail builds the HTML email for payment due date update
func (es *EmailService) buildPaymentDueUpdateEmail(data PaymentDueUpdateEmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #3b82f6 0%%, #2563eb 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #dbeafe; margin: 10px 0 0 0; font-size: 14px;">Payment Schedule Update</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    Review the update to your rent payment schedule for <strong>%s</strong>.
                </p>
                
                <div style="background-color: #eff6ff; border: 1px solid #bfdbfe; border-radius: 12px; padding: 20px; margin-bottom: 25px; text-align: center;">
                    <p style="color: #1e40af; font-size: 14px; margin: 0 0 10px 0; text-transform: uppercase; font-weight: 600;">New Payment Due Date</p>
                    <p style="color: #1e3a8a; font-size: 32px; margin: 0; font-weight: 700;">%s<span style="font-size: 16px; font-weight: 400; color: #60a5fa;"> of every month</span></p>
                </div>

                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 0 0 25px 0;">
                    Please ensure your future rent payments are made by this date to avoid any late fees.
                </p>

                <!-- Property Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1a365d; margin: 0 0 15px 0; font-size: 16px;">🏢 Property Details</h3>
                            <p style="margin: 5px 0;"><strong>Address:</strong> %s</p>
                        </td>
                    </tr>
                </table>


            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. Contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`,
		data.TenantName,
		data.PropertyTitle,
		data.NewDueDate,
		data.PropertyAddress,

	)
}

// RentUpdateEmailData contains data for rent update email
type RentUpdateEmailData struct {
	TenantName      string
	TenantEmail     string
	PropertyTitle   string
	PropertyAddress string
	OldRent         string
	NewRent         string
}

// SendRentUpdateEmail sends email to tenant for rent update
func (es *EmailService) SendRentUpdateEmail(data RentUpdateEmailData) error {
	subject := "📈 Rent Update Notice - Propbank"
	body := es.buildRentUpdateEmail(data)

	if data.TenantEmail != "" {
		return es.sendEmail(data.TenantEmail, subject, body)
	}
	return nil
}

// buildRentUpdateEmail builds the HTML email for rent update
func (es *EmailService) buildRentUpdateEmail(data RentUpdateEmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 14px;">Rent Update Notification</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    There has been an update to the monthly rent for your property <strong>%s</strong>.
                </p>
                
                <div style="background-color: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 12px; padding: 20px; margin-bottom: 25px;">
                    <table cellpadding="0" cellspacing="0" width="100%%">
                        <tr>
                            <td style="padding: 10px; border-bottom: 1px solid #bbf7d0;">
                                <strong style="color: #166534;">Previous Rent:</strong> 
                            </td>
                            <td style="padding: 10px; border-bottom: 1px solid #bbf7d0; text-align: right; color: #15803d; font-weight: 600;">
                                ₹%s
                            </td>
                        </tr>
                        <tr>
                            <td style="padding: 10px;">
                                <strong style="color: #166534;">New Rent:</strong> 
                            </td>
                            <td style="padding: 10px; text-align: right; color: #15803d; font-weight: 700; font-size: 18px;">
                                ₹%s
                            </td>
                        </tr>
                    </table>
                </div>

                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1a365d; margin: 0 0 15px 0; font-size: 16px;">🏢 Property Details</h3>
                            <p style="margin: 5px 0;"><strong>Address:</strong> %s</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
        <tr>
            <td style="background-color: #f8fafc; padding: 25px 30px; text-align: center; border-top: 1px solid #e2e8f0;">
                <p style="color: #64748b; font-size: 12px; margin: 0;">
                    This email was sent by Propbank. Contact us at 
                    <a href="mailto:connectpropbank@gmail.com" style="color: #2563eb;">connectpropbank@gmail.com</a>
                </p>
                <p style="color: #94a3b8; font-size: 11px; margin: 10px 0 0 0;">© 2025 Propbank. All rights reserved.</p>
            </td>
        </tr>
    </table>
</body>
</html>
`, data.TenantName, data.PropertyTitle, data.OldRent, data.NewRent, data.PropertyAddress)
}

// SendRenewalRequestNotification sends email to owner and tenant when tenant requests renewal
func (es *EmailService) SendRenewalRequestNotification(ownerEmail, ownerName, tenantName, tenantEmail, propertyTitle, propertyID string) error {
	// 1. Send Email to Owner
	ownerSubject := fmt.Sprintf("🔄 Renewal Request: %s", propertyTitle)
	ownerBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 14px;">Renewal Request</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    Your tenant, <strong>%s</strong>, has requested to renew the lease agreement for <strong>%s</strong>.
                </p>
                
                <div style="background-color: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 8px; padding: 20px; margin-bottom: 25px;">
                    <p style="color: #166534; font-size: 14px; margin: 0; font-weight: 500;">
                        Action Required: Please review this request and take appropriate action (Renew Agreement) through the Propbank dashboard.
                    </p>
                </div>

                <p style="color: #4a5568; font-size: 14px; line-height: 1.6; margin: 0;">
                    Property ID: %s
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, ownerName, tenantName, propertyTitle, propertyID)

	if err := es.sendEmail(ownerEmail, ownerSubject, ownerBody); err != nil {
		// Log but continue
	}

	// 2. Send Confirmation Email to Tenant
	if tenantEmail != "" {
		tenantSubject := "Renewal Request Received - Propbank"
		tenantBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5;">
    <table cellpadding="0" cellspacing="0" width="100%%" style="max-width: 600px; margin: 0 auto; background-color: #ffffff;">
        <tr>
            <td style="background: linear-gradient(135deg, #3b82f6 0%%, #2563eb 100%%); padding: 30px; text-align: center;">
                <h1 style="color: #ffffff; margin: 0; font-size: 28px;">🏠 Propbank</h1>
                <p style="color: #bfdbfe; margin: 10px 0 0 0; font-size: 14px;">Renewal Request Sent</p>
            </td>
        </tr>
        <tr>
            <td style="padding: 40px 30px;">
                <h2 style="color: #1a365d; margin: 0 0 20px 0; font-size: 22px;">Hi %s,</h2>
                
                <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 25px 0;">
                    We have received your request to renew the agreement for <strong>%s</strong>.
                </p>
                
                 <div style="background-color: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px; padding: 20px; margin-bottom: 25px;">
                     <p style="color: #1e40af; font-size: 14px; margin: 0;">
                        The owner (%s) has been notified. They will process your request shortly.
                    </p>
                </div>
            </td>
        </tr>
    </table>
</body>
</html>
`, tenantName, propertyTitle, ownerName)

		return es.sendEmail(tenantEmail, tenantSubject, tenantBody)
	}
	return nil
}
