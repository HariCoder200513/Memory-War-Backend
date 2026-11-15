package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/memory-wars/backend/internal/config"
	"github.com/memory-wars/backend/internal/queue"
)

type EmailWorker struct {
	sqsClient *sqs.Client
	queueURL  string
	cfg       *config.Config
}

func NewEmailWorker(sqsQueue *queue.SQSClient, cfg *config.Config) *EmailWorker {
	return &EmailWorker{
		sqsClient: sqsQueue.GetClient(),
		queueURL:  sqsQueue.GetQueueURL(),
		cfg:       cfg,
	}
}

func (w *EmailWorker) Start(ctx context.Context) {
	log.Println("Email worker started, polling SQS...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Email worker shutting down")
			return
		default:
			w.pollAndProcess(ctx)
		}
	}
}

func (w *EmailWorker) pollAndProcess(ctx context.Context) {
	result, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(w.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   30,
	})
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		log.Printf("Error receiving SQS messages: %v", err)
		time.Sleep(5 * time.Second)
		return
	}

	for _, msg := range result.Messages {
		var emailMsg queue.EmailMessage
		if err := json.Unmarshal([]byte(*msg.Body), &emailMsg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			w.deleteMessage(ctx, *msg.ReceiptHandle)
			continue
		}

		if err := w.sendEmail(emailMsg); err != nil {
			log.Printf("Failed to send email to %s: %v", emailMsg.To, err)
			continue
		}

		w.deleteMessage(ctx, *msg.ReceiptHandle)
		log.Printf("Email sent successfully to %s (type=%s)", emailMsg.To, emailMsg.Type)
	}
}

func (w *EmailWorker) deleteMessage(ctx context.Context, receiptHandle string) {
	_, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(w.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		log.Printf("Failed to delete SQS message: %v", err)
	}
}

func (w *EmailWorker) sendEmail(msg queue.EmailMessage) error {
	addr := fmt.Sprintf("%s:%d", w.cfg.SMTPHost, w.cfg.SMTPPort)
	auth := smtp.PlainAuth("", w.cfg.SMTPUser, w.cfg.SMTPPassword, w.cfg.SMTPHost)

	subject := msg.Subject
	body := w.buildEmailBody(msg)

	mime := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		w.cfg.SMTPFrom, msg.To, subject, body,
	)

	return smtp.SendMail(addr, auth, w.cfg.SMTPFrom, []string{msg.To}, []byte(mime))
}

func (w *EmailWorker) buildEmailBody(msg queue.EmailMessage) string {
	switch msg.Type {
	case "verification":
		return fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
				<h2 style="color: #1a1a2e;">Verify your email</h2>
				<p style="color: #555;">Your Memory Wars verification code is:</p>
				<div style="background: #f4f4f8; border-radius: 12px; padding: 24px; text-align: center; margin: 24px 0;">
					<span style="font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #1a1a2e;">%s</span>
				</div>
				<p style="color: #888; font-size: 13px;">This code expires in 10 minutes. If you didn't create an account, ignore this email.</p>
			</div>
		`, msg.Code)
	case "password_reset":
		return fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
				<h2 style="color: #1a1a2e;">Reset your password</h2>
				<p style="color: #555;">Your password reset code is:</p>
				<div style="background: #f4f4f8; border-radius: 12px; padding: 24px; text-align: center; margin: 24px 0;">
					<span style="font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #1a1a2e;">%s</span>
				</div>
				<p style="color: #888; font-size: 13px;">This code expires in 1 hour. If you didn't request a password reset, ignore this email.</p>
			</div>
		`, msg.Code)
	default:
		return fmt.Sprintf("<p>Your code: %s</p>", msg.Code)
	}
}
