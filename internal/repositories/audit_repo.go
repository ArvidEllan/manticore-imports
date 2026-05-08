package repositories

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"manticore-imports/internal/domain"
)

type AuditRepository struct {
	client *dynamodb.Client
	table  string
}

func NewAuditRepository(client *dynamodb.Client, table string) *AuditRepository {
	return &AuditRepository{client: client, table: table}
}

func (r *AuditRepository) Create(ctx context.Context, event domain.AuditEvent) error {
	av, err := attributevalue.MarshalMap(event)
	if err != nil { return err }
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.table), Item: av})
	return err
}
