package repositories

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"manticore-imports/internal/domain"
)

type DocumentRepository struct {
	client *dynamodb.Client
	table  string
}

func NewDocumentRepository(client *dynamodb.Client, table string) *DocumentRepository {
	return &DocumentRepository{client: client, table: table}
}

func (r *DocumentRepository) Create(ctx context.Context, item domain.DocumentRecord) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil { return err }
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.table), Item: av})
	return err
}
