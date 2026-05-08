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
	if s.from == "" || to == "" { return nil }
	_, err := s.client.SendEmail(ctx, &ses.SendEmailInput{
		Source: aws.String(s.from),
		Destination: &stypes.Destination{ToAddresses: []string{to}},
		Message: &stypes.Message{
			Subject: &stypes.Content{Data: aws.String(subject)},
			Body: &stypes.Body{Text: &stypes.Content{Data: aws.String(body)}},
		},
	})
	if err != nil { return fmt.Errorf("send email: %w", err) }
	return nil
}
