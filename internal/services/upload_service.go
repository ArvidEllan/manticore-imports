package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type UploadService struct {
	presignClient *s3.PresignClient
	bucket        string
}

func NewUploadService(client *s3.Client, bucket string) *UploadService {
	return &UploadService{presignClient: s3.NewPresignClient(client), bucket: bucket}
}

func (s *UploadService) CreatePresignedUpload(ctx context.Context, requestID, fileName, contentType string) (string, string, error) {
	docID := uuid.NewString()
	key := fmt.Sprintf("requests/%s/documents/%s/%s", requestID, docID, fileName)
	resp, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		ContentType: &contentType,
		ACL:         types.ObjectCannedACLPrivate,
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil { return "", "", err }
	return docID, resp.URL, nil
}
