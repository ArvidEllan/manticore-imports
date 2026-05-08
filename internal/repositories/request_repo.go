package repositories

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"manticore-imports/internal/domain"
)

type RequestRepository struct {
	client *dynamodb.Client
	table  string
}

func NewRequestRepository(client *dynamodb.Client, table string) *RequestRepository {
	return &RequestRepository{client: client, table: table}
}

func (r *RequestRepository) Create(ctx context.Context, item domain.Request) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil { return err }
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.table), Item: av})
	return err
}

func (r *RequestRepository) GetByReference(ctx context.Context, reference string) (*domain.Request, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String("reference-index"),
		KeyConditionExpression: aws.String("reference = :reference"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":reference": &ddbtypes.AttributeValueMemberS{Value: reference},
		},
		Limit: aws.Int32(1),
	})
	if err != nil { return nil, err }
	if len(out.Items) == 0 { return nil, nil }
	var item domain.Request
	if err := attributevalue.UnmarshalMap(out.Items[0], &item); err != nil { return nil, err }
	return &item, nil
}

func (r *RequestRepository) GetByID(ctx context.Context, requestID string) (*domain.Request, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.table),
		IndexName:              aws.String("request-id-index"),
		KeyConditionExpression: aws.String("requestId = :requestId"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":requestId": &ddbtypes.AttributeValueMemberS{Value: requestID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil { return nil, err }
	if len(out.Items) == 0 { return nil, nil }
	var item domain.Request
	if err := attributevalue.UnmarshalMap(out.Items[0], &item); err != nil { return nil, err }
	return &item, nil
}

func (r *RequestRepository) List(ctx context.Context) ([]domain.Request, error) {
	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{TableName: aws.String(r.table)})
	if err != nil { return nil, err }
	items := make([]domain.Request, 0, len(out.Items))
	for _, av := range out.Items {
		var req domain.Request
		if err := attributevalue.UnmarshalMap(av, &req); err != nil { return nil, err }
		items = append(items, req)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (r *RequestRepository) UpdateStatus(ctx context.Context, requestID, status string, updatedAt string) error {
	item, err := r.GetByID(ctx, requestID)
	if err != nil { return err }
	if item == nil { return fmt.Errorf("request not found") }
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.table),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: item.PK},
		},
		UpdateExpression: aws.String("SET #status = :status, updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]string{"#status": "status"},
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":status": &ddbtypes.AttributeValueMemberS{Value: status},
			":updatedAt": &ddbtypes.AttributeValueMemberS{Value: updatedAt},
		},
	})
	return err
}
