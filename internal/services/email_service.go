package services

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	stypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailService struct {
	client *ses.Client
	from   string
}

func NewEmailService(client *ses.Client, from string) *EmailService {
	return &EmailService{client: client, from: from}
}

func (s *EmailService) Send(ctx context.Context, to, subject, body string) error {
	return s.SendHTML(ctx, to, subject, "", body)
}

func (s *EmailService) SendHTML(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if s.from == "" || to == "" {
		return nil
	}
	msg := &stypes.Message{
		Subject: &stypes.Content{Data: aws.String(subject)},
		Body:    &stypes.Body{},
	}
	if htmlBody != "" {
		msg.Body.Html = &stypes.Content{Data: aws.String(htmlBody)}
	}
	if textBody != "" {
		msg.Body.Text = &stypes.Content{Data: aws.String(textBody)}
	}
	if msg.Body.Html == nil && msg.Body.Text == nil {
		return fmt.Errorf("email body is empty")
	}
	_, err := s.client.SendEmail(ctx, &ses.SendEmailInput{
		Source:      aws.String(s.from),
		Destination: &stypes.Destination{ToAddresses: []string{to}},
		Message:     msg,
	})
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
