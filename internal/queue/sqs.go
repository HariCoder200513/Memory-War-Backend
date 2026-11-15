package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Code    string `json:"code"`
	Type    string `json:"type"`
}

type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSClient(region, queueURL string) (*SQSClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &SQSClient{
		client:   sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}

func (s *SQSClient) PublishEmailMessage(msg EmailMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = s.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("failed to send SQS message: %w", err)
	}

	log.Printf("Published email message to SQS: type=%s to=%s", msg.Type, msg.To)
	return nil
}

func (s *SQSClient) ReceiveMessages(ctx context.Context) ([]sqs.ReceiveMessageOutput, error) {
	result, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}
	_ = result
	return nil, nil
}

func (s *SQSClient) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

func (s *SQSClient) GetClient() *sqs.Client {
	return s.client
}

func (s *SQSClient) GetQueueURL() string {
	return s.queueURL
}
