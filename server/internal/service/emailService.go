package service

import (
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"semki/pkg/telemetry"
)

type EmailService struct {
	Dialer   *gomail.Dialer
	From     string
	FromName string
}

func NewEmailService(host string, port int, username, password, from, fromName string) *EmailService {
	dialer := gomail.NewDialer(host, port, username, password)

	return &EmailService{
		Dialer:   dialer,
		From:     from,
		FromName: fromName,
	}
}

const verificationEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6; 
            color: #333; 
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
        }
        .container { 
            max-width: 600px; 
            margin: 40px auto; 
            background: white;
            padding: 40px; 
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        h2 {
            color: #1a1a1a;
            margin-top: 0;
        }
        .button { 
            display: inline-block; 
            padding: 14px 32px; 
            background-color: #0066ff; 
            color: #ffffff !important; 
            text-decoration: none; 
            border-radius: 6px; 
            margin: 24px 0;
            font-weight: 600;
            font-size: 16px;
        }
        .button:hover {
            background-color: #0052cc;
        }
        .link-box {
            background: #f8f9fa;
            padding: 16px;
            border-radius: 6px;
            word-break: break-all;
            font-size: 13px;
            color: #666;
            border: 1px solid #e1e4e8;
        }
        .footer { 
            margin-top: 32px; 
            padding-top: 24px;
            border-top: 1px solid #eee;
            font-size: 13px; 
            color: #666; 
        }
        .warning {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 12px 16px;
            margin: 20px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Welcome, {{.Name}}!</h2>
        <p>Thank you for signing up. Please verify your email address to activate your account.</p>
        
        <div style="text-align: center;">
            <a href="{{.VerificationLink}}" class="button">Verify Email Address</a>
        </div>
        
        <p>Or copy and paste this link into your browser:</p>
        <div class="link-box">{{.VerificationLink}}</div>
        
        <div class="warning">
            <strong>⏰ This link expires in 24 hours</strong>
        </div>
        
        <div class="footer">
            <p>If you didn't create an account, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`

const invitationEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6; 
            color: #333; 
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
        }
        .container { 
            max-width: 600px; 
            margin: 40px auto; 
            background: white;
            padding: 40px; 
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        h2 {
            color: #1a1a1a;
            margin-top: 0;
        }
        .org-name {
            color: #0066ff;
            font-weight: 600;
        }
        .button { 
            display: inline-block; 
            padding: 14px 32px; 
            background-color: #28a745; 
            color: #ffffff !important; 
            text-decoration: none; 
            border-radius: 6px; 
            margin: 24px 0;
            font-weight: 600;
            font-size: 16px;
        }
        .button:hover {
            background-color: #218838;
        }
        .link-box {
            background: #f8f9fa;
            padding: 16px;
            border-radius: 6px;
            word-break: break-all;
            font-size: 13px;
            color: #666;
            border: 1px solid #e1e4e8;
        }
        .footer { 
            margin-top: 32px; 
            padding-top: 24px;
            border-top: 1px solid #eee;
            font-size: 13px; 
            color: #666; 
        }
        .info-box {
            background: #e7f3ff;
            border-left: 4px solid #0066ff;
            padding: 16px;
            margin: 20px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>You've been invited!</h2>
        <p>Hi {{.Name}},</p>
        <p>You have been invited to join <span class="org-name">{{.OrganizationName}}</span>.</p>
        
        <div class="info-box">
            <strong>What's next?</strong><br>
            Click the button below to accept your invitation and set up your account.
        </div>
        
        <div style="text-align: center;">
            <a href="{{.InviteLink}}" class="button">Accept Invitation</a>
        </div>
        
        <p>Or copy and paste this link:</p>
        <div class="link-box">{{.InviteLink}}</div>
        
        <div class="footer">
            <p><strong>This invitation expires in 7 days.</strong></p>
            <p>If you weren't expecting this invitation, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6; 
            color: #333; 
            background-color: #f5f5f5;
            margin: 0;
            padding: 0;
        }
        .container { 
            max-width: 600px; 
            margin: 40px auto; 
            background: white;
            padding: 40px; 
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        h2 {
            color: #1a1a1a;
            margin-top: 0;
        }
        .button { 
            display: inline-block; 
            padding: 14px 32px; 
            background-color: #dc3545; 
            color: #ffffff !important; 
            text-decoration: none; 
            border-radius: 6px; 
            margin: 24px 0;
            font-weight: 600;
            font-size: 16px;
        }
        .button:hover {
            background-color: #c82333;
        }
        .link-box {
            background: #f8f9fa;
            padding: 16px;
            border-radius: 6px;
            word-break: break-all;
            font-size: 13px;
            color: #666;
            border: 1px solid #e1e4e8;
        }
        .footer { 
            margin-top: 32px; 
            padding-top: 24px;
            border-top: 1px solid #eee;
            font-size: 13px; 
            color: #666; 
        }
        .warning {
            background: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 12px 16px;
            margin: 20px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Password Reset Request</h2>
        <p>Hi {{.Name}},</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        
        <div style="text-align: center;">
            <a href="{{.ResetLink}}" class="button">Reset Password</a>
        </div>
        
        <p>Or copy and paste this link:</p>
        <div class="link-box">{{.ResetLink}}</div>
        
        <div class="warning">
            <strong>⏰ This link expires in 1 hour</strong>
        </div>
        
        <div class="footer">
            <p><strong>Didn't request a password reset?</strong></p>
            <p>If you didn't request this, please ignore this email. Your password will remain unchanged.</p>
        </div>
    </div>
</body>
</html>
`

type VerificationEmailData struct {
	Name             string
	VerificationLink string
}

type InvitationEmailData struct {
	Name             string
	OrganizationName string
	InviteLink       string
}

type PasswordResetEmailData struct {
	Name      string
	ResetLink string
}

func (e *EmailService) SendVerificationEmail(toEmail, name, verificationLink string) error {
	tmpl, err := template.New("verification").Parse(verificationEmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	data := VerificationEmailData{
		Name:             name,
		VerificationLink: verificationLink,
	}

	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.From, e.FromName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Verify your email address")
	m.SetBody("text/html", body.String())

	plainText := fmt.Sprintf("Hi %s,\n\nPlease verify your email: %s\n\nThis link expires in 24 hours.",
		name, verificationLink)
	m.AddAlternative("text/plain", plainText)

	telemetry.Log.Info(fmt.Sprintf("Sending verification email to: %s", toEmail))
	if err := e.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	telemetry.Log.Info(fmt.Sprintf("Verification email sent to: %s", toEmail))
	return nil
}

func (e *EmailService) SendInvitationEmail(toEmail, name, organizationName, inviteLink string) error {
	tmpl, err := template.New("invitation").Parse(invitationEmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	data := InvitationEmailData{
		Name:             name,
		OrganizationName: organizationName,
		InviteLink:       inviteLink,
	}

	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.From, e.FromName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("Invitation to join %s", organizationName))
	m.SetBody("text/html", body.String())

	plainText := fmt.Sprintf("Hi %s,\n\nYou've been invited to join %s.\n\nAccept invitation: %s\n\nExpires in 7 days.",
		name, organizationName, inviteLink)
	m.AddAlternative("text/plain", plainText)

	telemetry.Log.Info(fmt.Sprintf("Sending invitation email to: %s", toEmail))
	if err := e.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	telemetry.Log.Info(fmt.Sprintf("Invitation email sent to: %s", toEmail))
	return nil
}

func (e *EmailService) SendPasswordResetEmail(toEmail, name, resetLink string) error {
	tmpl, err := template.New("password_reset").Parse(passwordResetEmailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	data := PasswordResetEmailData{
		Name:      name,
		ResetLink: resetLink,
	}

	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.From, e.FromName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Password Reset Request")
	m.SetBody("text/html", body.String())

	plainText := fmt.Sprintf("Hi %s,\n\nReset your password: %s\n\nExpires in 1 hour.",
		name, resetLink)
	m.AddAlternative("text/plain", plainText)

	telemetry.Log.Info(fmt.Sprintf("Sending password reset email to: %s", toEmail))
	if err := e.Dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	telemetry.Log.Info(fmt.Sprintf("Password reset email sent to: %s", toEmail))
	return nil
}
