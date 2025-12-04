package services

import (
	"context"
	"time"
)

// ReminderScheduler handles scheduling and sending visit reminders
type ReminderScheduler struct {
	emailService *EmailService
	visitService *VisitService
	userService  *UserService
	ticker       *time.Ticker
	stopChan     chan bool
	isRunning    bool
}

// NewReminderScheduler creates a new reminder scheduler
func NewReminderScheduler(emailService *EmailService, visitService *VisitService, userService *UserService) *ReminderScheduler {
	return &ReminderScheduler{
		emailService: emailService,
		visitService: visitService,
		userService:  userService,
		stopChan:     make(chan bool),
		isRunning:    false,
	}
}

// Start starts the reminder scheduler that runs every minute
func (rs *ReminderScheduler) Start() {
	if rs.isRunning {
		return
	}

	rs.isRunning = true
	rs.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		// Run immediately on start
		rs.checkAndSendReminders()

		for {
			select {
			case <-rs.ticker.C:
				rs.checkAndSendReminders()
			case <-rs.stopChan:
				rs.ticker.Stop()
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

// ForceCheck forces an immediate check for reminders (useful for testing)
func (rs *ReminderScheduler) ForceCheck() {
	rs.checkAndSendReminders()
}
