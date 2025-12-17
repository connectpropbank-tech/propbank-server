package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"shoprop-backend/models"
	"time"
)

// India Standard Time (IST) is UTC+5:30
var IST = time.FixedZone("IST", 5*60*60+30*60)

type EmailService struct {
	SMTPHost string
	SMTPPort string
	Email    string
	Password string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost: getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort: getEnvOrDefault("SMTP_PORT", "587"),
		Email:    getEnvOrDefault("SMTP_EMAIL", "connectpropbank@gmail.com"),
		Password: os.Getenv("SMTP_PASSWORD"), // Gmail App Password - required
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

// sendEmail sends an email using SMTP with TLS
func (es *EmailService) sendEmail(to, subject, body string) error {

	// Check if email credentials are properly configured
	if es.Password == "" {
		return nil
	}

	// Build email message with proper headers
	from := es.Email
	msg := fmt.Sprintf("From: Propbank <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n%s", from, to, subject, body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", es.SMTPHost, es.SMTPPort)
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: es.SMTPHost,
	}
	if err = conn.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", es.Email, es.Password, es.SMTPHost)
	if err = conn.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// Set sender
	if err = conn.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// Set recipient
	if err = conn.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send message body
	writer, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

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
	// Send email to Tenant
	tenantSubject := "🏠 Welcome! You've been added as a Tenant - Propbank"
	tenantBody := es.buildTenantAddedEmail(data, true)

	if data.TenantEmail != "" {
		if err := es.sendEmail(data.TenantEmail, tenantSubject, tenantBody); err != nil {
		} else {
		}
	}

	// Send email to Owner
	ownerSubject := "🏠 Tenant Added to Your Property - Propbank"
	ownerBody := es.buildTenantAddedEmail(data, false)

	if data.OwnerEmail != "" {
		if err := es.sendEmail(data.OwnerEmail, ownerSubject, ownerBody); err != nil {
			return err
		}
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

                <!-- Tenant Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #eff6ff; border-radius: 12px; border: 1px solid #bfdbfe; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1e40af; margin: 0 0 15px 0; font-size: 16px;">👤 Tenant Details</h3>
                            <p style="margin: 5px 0;"><strong>Name:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Email:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Phone:</strong> %s</p>
                        </td>
                    </tr>
                </table>

                <!-- Owner Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #fef3c7; border-radius: 12px; border: 1px solid #fcd34d; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #92400e; margin: 0 0 15px 0; font-size: 16px;">🏠 Owner Details</h3>
                            <p style="margin: 5px 0;"><strong>Name:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Email:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Phone:</strong> %s</p>
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
		data.TenantName,
		data.TenantEmail,
		data.TenantPhone,
		data.OwnerName,
		data.OwnerEmail,
		data.OwnerPhone,
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
	TerminationDate    string
	Reason             string
	AgreementStartDate string
	AgreementEndDate   string
	AgreementPeriod    string
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

// buildAgreementTerminationEmail builds the HTML email for agreement termination notification
func (es *EmailService) buildAgreementTerminationEmail(data AgreementTerminationEmailData, isForTenant bool) string {
	recipientName := data.OwnerName
	headerText := "Agreement Terminated"
	headerColor := "#059669"
	introText := "The rental agreement has been successfully terminated."

	if isForTenant {
		recipientName = data.TenantName
		headerText = "Agreement Termination Notice"
		headerColor := "#dc2626"
		introText = "The property owner has terminated your rental agreement. Please review the details below."
		_ = headerColor // avoid unused variable
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

                <!-- Tenant Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #eff6ff; border-radius: 12px; border: 1px solid #bfdbfe; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #1e40af; margin: 0 0 15px 0; font-size: 16px;">👤 Tenant Details</h3>
                            <p style="margin: 5px 0;"><strong>Name:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Email:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Phone:</strong> %s</p>
                        </td>
                    </tr>
                </table>

                <!-- Owner Details -->
                <table cellpadding="0" cellspacing="0" width="100%%" style="background-color: #fef3c7; border-radius: 12px; border: 1px solid #fcd34d; margin-bottom: 20px;">
                    <tr>
                        <td style="padding: 20px;">
                            <h3 style="color: #92400e; margin: 0 0 15px 0; font-size: 16px;">🏠 Owner Details</h3>
                            <p style="margin: 5px 0;"><strong>Name:</strong> %s</p>
                            <p style="margin: 5px 0;"><strong>Email:</strong> <a href="mailto:%s" style="color: #2563eb;">%s</a></p>
                            <p style="margin: 5px 0;"><strong>Phone:</strong> %s</p>
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
		data.PropertyTitle,
		data.PropertyAddress,
		data.PropertyType,
		data.TenantName,
		data.TenantEmail,
		data.TenantPhone,
		data.OwnerName,
		data.OwnerEmail,
		data.OwnerEmail,
		data.OwnerPhone,
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
