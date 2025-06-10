package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"
)

// EmailData represents the data structure for email template you could add more fields as needed
type EmailData struct {
	VerificationCode string
	ResetLink        string
}

// sendVerificationEmail sends a verification code email to the specified address
func sendVerificationEmail(toEmail, verificationCode string) error {
	// Parse the HTML template
	tmpl, err := template.ParseFiles("templates/verification_code.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Prepare template data
	data := EmailData{
		VerificationCode: verificationCode,
	}

	// Execute template with data
	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Create the email message
	message := mg.NewMessage(
		"noreply@"+mg.Domain(), // From address
		"Verify Your Account",  // Subject
		htmlBody.String(),      // since we're using HTML
		toEmail,                // To address
	)

	// Set HTML body
	message.SetHTML(htmlBody.String())

	// Send the email with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err = mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email via Mailgun: %w", err)
	}

	return nil
}

// sendPasswordResetEmail sends a password reset link email to the specified address
func sendPasswordResetEmail(toEmail, resetToken string) error {
	// Parse the HTML template
	tmpl, err := template.ParseFiles("templates/password_reset.html")
	if err != nil {
		return fmt.Errorf("failed to parse password reset template: %w", err)
	}

	// Construct the reset link (adjust the domain as needed)
	resetLink := fmt.Sprintf("http://localhost:8080/reset?token=%s", resetToken)

	// Prepare template data
	data := EmailData{
		ResetLink: resetLink,
	}

	// Execute template with data
	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to execute password reset template: %w", err)
	}

	// Create the email message
	message := mg.NewMessage(
		"noreply@"+mg.Domain(), // From address
		"Reset Your Password",  // Subject
		htmlBody.String(),      // we're using HTML
		toEmail,                // To address
	)

	// Set HTML body
	message.SetHTML(htmlBody.String())

	// Send the email with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, _, err = mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send password reset email via Mailgun: %w", err)
	}

	return nil
}
