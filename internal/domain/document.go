package domain

import "time"

type DocumentRecord struct {
	PK          string    `dynamodbav:"pk" json:"-"`
	SK          string    `dynamodbav:"sk" json:"-"`
	DocumentID  string    `dynamodbav:"documentId" json:"documentId"`
	RequestID   string    `dynamodbav:"requestId" json:"requestId"`
	FileName    string    `dynamodbav:"fileName" json:"fileName"`
	ContentType string    `dynamodbav:"contentType" json:"contentType"`
	S3Key       string    `dynamodbav:"s3Key" json:"s3Key"`
	CreatedAt   time.Time `dynamodbav:"createdAt" json:"createdAt"`
}

type AuditEvent struct {
	PK        string    `dynamodbav:"pk" json:"-"`
	SK        string    `dynamodbav:"sk" json:"-"`
	EventType string    `dynamodbav:"eventType" json:"eventType"`
	Actor     string    `dynamodbav:"actor" json:"actor"`
	Details   string    `dynamodbav:"details" json:"details"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
}
