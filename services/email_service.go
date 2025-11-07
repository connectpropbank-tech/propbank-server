package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"shoprop-backend/models"
	"time"
)

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
		Email:    getEnvOrDefault("SMTP_EMAIL", "your-email@gmail.com"),
		Password: getEnvOrDefault("SMTP_PASSWORD", "your-app-password"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SendVisitReminder sends an email reminder for a scheduled visit
func (es *EmailService) SendVisitReminder(user *models.User, visit *models.Visit) error {
	subject := fmt.Sprintf("Reminder: Property Visit - %s", visit.Title)

	body := fmt.Sprintf(`
Dear %s,

This is a friendly reminder about your upcoming property visit:

📍 Visit: %s
📅 Date & Time: %s
📝 Details: %s

Please make sure to:
- Arrive on time for your scheduled visit
- Bring necessary documents (ID, income proof if required)
- Prepare any questions you want to ask about the property

If you need to reschedule or cancel this visit, please log in to your ShoPROP account.

Best regards,
ShoPROP Team

---
This is an automated reminder. Please do not reply to this email.
`,
		user.Name,
		visit.Title,
		visit.VisitDate.Format("Monday, January 2, 2006 at 3:04 PM"),
		visit.Description,
	)

	return es.sendEmail(user.Email, subject, body)
}

// SendVisitConfirmation sends a confirmation email when a visit is scheduled
func (es *EmailService) SendVisitConfirmation(user *models.User, visit *models.Visit) error {
	subject := fmt.Sprintf("Visit Scheduled: %s", visit.Title)

	body := fmt.Sprintf(`
Dear %s,

Your property visit has been successfully scheduled!

📍 Visit: %s
📅 Date & Time: %s
📝 Details: %s
🔔 Reminder: You'll receive a reminder %d minutes before the visit via %s

Visit Details:
- Make sure to arrive on time
- Bring necessary identification
- Prepare questions about the property

You can manage your visits by logging into your ShoPROP account.

Best regards,
ShoPROP Team

---
This is an automated confirmation. Please do not reply to this email.
`,
		user.Name,
		visit.Title,
		visit.VisitDate.Format("Monday, January 2, 2006 at 3:04 PM"),
		visit.Description,
		visit.ReminderTime,
		visit.ReminderType,
	)

	return es.sendEmail(user.Email, subject, body)
}

// sendEmail sends an email using SMTP
func (es *EmailService) sendEmail(to, subject, body string) error {
	log.Printf("🔧 Email Service Config - Host: %s, Port: %s, From: %s", es.SMTPHost, es.SMTPPort, es.Email)

	// Check if email credentials are properly configured
	if es.Email == "your-email@gmail.com" || es.Password == "your-app-password" || es.Email == "shoprop.notifications@gmail.com" {
		log.Printf("📧 Email simulation mode: Would send to %s with subject: %s", to, subject)
		log.Printf("📧 Email body preview: %s", body[:min(100, len(body))])
		log.Printf("⚠️ To enable real email sending, configure SMTP_EMAIL and SMTP_PASSWORD in .env file")
		return nil
	}

	auth := smtp.PlainAuth("", es.Email, es.Password, es.SMTPHost)

	// Properly format email headers
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", es.Email, to, subject, body)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", es.SMTPHost, es.SMTPPort),
		auth,
		es.Email,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("✅ Email sent successfully to %s", to)
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
	log.Println("🔔 Checking for visit reminders...")

	// This would typically be called by a cron job or background service
	// For now, it's a placeholder for the reminder system
	visits, err := visitService.GetPendingReminders()
	if err != nil {
		log.Printf("❌ Error getting pending reminders: %v", err)
		return
	}

	ctx := context.Background()
	for _, visit := range visits {
		user, err := userService.GetUserByID(ctx, visit.UserID)
		if err != nil {
			log.Printf("❌ Error getting user for visit reminder: %v", err)
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
