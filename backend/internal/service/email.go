package service

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/resend/resend-go/v2"
)

// EmailSender is the interface for email sending strategies
type EmailSender interface {
	Send(to, subject, html, text string) error
}

// SMTPSender sends emails via SMTP (for development with Mailpit)
type SMTPSender struct {
	host      string
	port      string
	fromEmail string
}

// NewSMTPSender creates a new SMTP sender
func NewSMTPSender(host, port, fromEmail string) *SMTPSender {
	return &SMTPSender{
		host:      host,
		port:      port,
		fromEmail: fromEmail,
	}
}

// Send sends an email via SMTP
func (s *SMTPSender) Send(to, subject, html, text string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	// Extract email from "Name <email>" format
	from := s.fromEmail
	if idx := strings.Index(from, "<"); idx != -1 {
		from = strings.TrimSuffix(from[idx+1:], ">")
	}

	// Build the email message with proper headers
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		s.fromEmail, to, subject, html,
	))

	// Send without authentication (Mailpit doesn't require it)
	return smtp.SendMail(addr, nil, from, []string{to}, msg)
}

// ResendSender sends emails via Resend API (for production)
type ResendSender struct {
	client    *resend.Client
	fromEmail string
}

// NewResendSender creates a new Resend sender
func NewResendSender(apiKey, fromEmail string) *ResendSender {
	return &ResendSender{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
}

// Send sends an email via Resend API
func (s *ResendSender) Send(to, subject, html, text string) error {
	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// NoopSender does nothing (for when no email service is configured)
type NoopSender struct{}

// Send does nothing and returns nil
func (s *NoopSender) Send(to, subject, html, text string) error {
	// Log for debugging
	fmt.Printf("[NoopSender] Would send email to: %s, subject: %s\n", to, subject)
	return nil
}

// EmailService handles email sending operations
type EmailService struct {
	sender EmailSender
	appURL string
}

// NewEmailService creates a new email service with the appropriate sender strategy
func NewEmailService() *EmailService {
	fromEmail := os.Getenv("EMAIL_FROM")
	if fromEmail == "" {
		fromEmail = "Team Todo <noreply@example.com>"
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	// Select the appropriate sender strategy
	sender := createSender(fromEmail)

	return &EmailService{
		sender: sender,
		appURL: appURL,
	}
}

// createSender creates the appropriate EmailSender based on environment variables
func createSender(fromEmail string) EmailSender {
	// Strategy 1: SMTP (for development with Mailpit)
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		smtpPort := os.Getenv("SMTP_PORT")
		if smtpPort == "" {
			smtpPort = "1025"
		}
		fmt.Println("[EmailService] Using SMTP sender (Mailpit)")
		return NewSMTPSender(smtpHost, smtpPort, fromEmail)
	}

	// Strategy 2: Resend API (for production)
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey != "" {
		fmt.Println("[EmailService] Using Resend sender")
		return NewResendSender(resendAPIKey, fromEmail)
	}

	// Strategy 3: Noop (no email service configured)
	fmt.Println("[EmailService] No email sender configured, using NoopSender")
	return &NoopSender{}
}

// SendInviteEmail sends an invitation email to join an organization
func (s *EmailService) SendInviteEmail(ctx context.Context, toEmail, inviterName, orgName, token string) error {
	inviteURL := fmt.Sprintf("%s/invite/%s", s.appURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Team Todoへの招待</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">Team Todo</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <h2 style="color: #333; margin-top: 0;">%s さんから招待が届いています</h2>
        <p>%s から「<strong>%s</strong>」への参加招待が届きました。</p>
        <p>以下のボタンをクリックして参加してください：</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block;">招待を承認する</a>
        </div>
        <p style="color: #666; font-size: 14px;">このリンクは7日間有効です。</p>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">
            このメールに心当たりがない場合は、無視していただいて構いません。<br>
            リンクが機能しない場合は、以下のURLをブラウザに貼り付けてください：<br>
            <a href="%s" style="color: #667eea;">%s</a>
        </p>
    </div>
</body>
</html>
`, inviterName, inviterName, orgName, inviteURL, inviteURL, inviteURL)

	text := fmt.Sprintf(`
%s さんから招待が届いています

%s から「%s」への参加招待が届きました。

以下のリンクをクリックして参加してください：
%s

このリンクは7日間有効です。

このメールに心当たりがない場合は、無視していただいて構いません。
`, inviterName, inviterName, orgName, inviteURL)

	subject := fmt.Sprintf("[Team Todo] %s から「%s」への招待", inviterName, orgName)
	return s.sender.Send(toEmail, subject, html, text)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail, displayName string) error {
	loginURL := fmt.Sprintf("%s/login", s.appURL)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Team Todoへようこそ</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">Team Todo</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <h2 style="color: #333; margin-top: 0;">%s さん、ようこそ！</h2>
        <p>Team Todoへのご登録ありがとうございます。</p>
        <p>チームのタスク管理を効率的に行うために、Team Todoをご活用ください。</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; display: inline-block;">ログインする</a>
        </div>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        <p style="color: #999; font-size: 12px;">
            ご不明な点がございましたら、お気軽にお問い合わせください。
        </p>
    </div>
</body>
</html>
`, displayName, loginURL)

	text := fmt.Sprintf(`
%s さん、ようこそ！

Team Todoへのご登録ありがとうございます。
チームのタスク管理を効率的に行うために、Team Todoをご活用ください。

ログインはこちらから：
%s

ご不明な点がございましたら、お気軽にお問い合わせください。
`, displayName, loginURL)

	return s.sender.Send(toEmail, "[Team Todo] ご登録ありがとうございます", html, text)
}
