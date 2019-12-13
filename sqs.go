package theta

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/phogolabs/log"
)

// SQSHandler dispatches the commands
type SQSHandler struct {
	EventHandler EventHandler
}

// HandleContext dispatches the command to the underlying executors
func (r *SQSHandler) HandleContext(ctx context.Context, input events.SQSEvent) error {
	for _, record := range input.Records {
		logger := log.WithFields(
			log.Map{
				"sqs_message_id":       record.MessageId,
				"sqs_event_source":     record.EventSource,
				"sqs_event_source_arn": record.EventSourceARN,
				"sqs_receipt_handle":   record.ReceiptHandle,
			},
		)

		args := &EventArgs{}

		reader := base64.NewDecoder(
			base64.StdEncoding,
			bytes.NewBufferString(record.Body),
		)

		logger.Info("unmarshaling event")
		if err := json.NewDecoder(reader).Decode(args); err != nil {
			logger.WithError(err).Error("failed to unmarshal event")
			return err
		}

		logger = log.WithFields(
			log.Map{
				"event_id":     args.Event.ID,
				"event_name":   args.Event.Name,
				"event_sender": args.Event.Sender,
				"event_source": args.Event.Source,
			},
		)

		ctx = log.SetContext(ctx, logger)

		logger.Info("validating event")
		if err := validation.StructCtx(ctx, args); err != nil {
			logger.WithError(err).Error("failed to validate event")
			return err
		}

		logger.Info("handling event")
		if err := r.EventHandler.HandleContext(ctx, args); err != nil {
			logger.WithError(err).Error("failed to execute args")
			return err
		}
	}

	return nil
}

//go:generate counterfeiter -fake-name SQSClient -o ./fake/sqs_client.go . SQSClient

// SQSClient creates a new client
type SQSClient = sqsiface.SQSAPI

var _ EventHandler = &SQSDispatcher{}

// SQSDispatcher dispatches a command via SQS
type SQSDispatcher struct {
	// QueueURL address
	QueueURL string
	// Clinet for the SQS
	Client SQSClient
}

// HandleContext dispatches an event to SQS
func (r *SQSDispatcher) HandleContext(ctx context.Context, args *EventArgs) error {
	logger := log.WithFields(
		log.Map{
			"event_id":     args.Event.ID,
			"event_name":   args.Event.Name,
			"event_sender": args.Event.Sender,
			"event_source": args.Event.Source,
			"event_target": r.QueueURL,
		},
	)

	logger.Info("marshaling event")
	body, err := json.Marshal(args)
	if err != nil {
		logger.WithError(err).Error("failed to marshal event")
		return err
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(r.QueueURL),
		MessageBody: aws.String(base64.StdEncoding.EncodeToString(body)),
	}

	logger.Info("dispatching event to sqs")
	_, err = r.Client.SendMessageWithContext(ctx, input)
	return err
}
