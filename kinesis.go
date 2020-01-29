package theta

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	consumer "github.com/harlow/kinesis-consumer"
	"github.com/phogolabs/log"
)

// KinesisHandler reacts on events
type KinesisHandler struct {
	EventHandler EventHandler
}

// HandleContext dispatches the event to the handler
func (r *KinesisHandler) HandleContext(ctx context.Context, input events.KinesisEvent) error {
	for index, record := range input.Records {
		logger := log.WithFields(
			log.Map{
				"kinesis_record_index":     index,
				"kinesis_event_id":         record.EventID,
				"kinesis_event_name":       record.EventName,
				"kinesis_event_source":     record.EventSource,
				"kinesis_event_source_arn": record.EventSourceArn,
				"kinesis_partition_key":    record.Kinesis.PartitionKey,
				"kinesis_sequence_number":  record.Kinesis.SequenceNumber,
			},
		)

		args := &EventArgs{}

		logger.Info("unmarshaling event")
		if err := json.Unmarshal(record.Kinesis.Data, args); err != nil {
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
			logger.WithError(err).Error("failed to handle event")
			return err
		}
	}

	return nil
}

//go:generate counterfeiter -fake-name KinesisClient -o ./fake/kinesis_client.go . KinesisClient

// KinesisClient creates a new client
type KinesisClient = kinesisiface.KinesisAPI

// KinesisDispatcherConfig represents the kinesis dispatcher config
type KinesisDispatcherConfig struct {
	RoleArn    string
	Region     string
	StreamName string
}

var _ EventHandler = &KinesisDispatcher{}

// KinesisDispatcher dispatches event to kinesis bus
type KinesisDispatcher struct {
	StreamName string
	Client     KinesisClient
}

// NewKinesisDispatcher creates a new dispatcher to kinesis
func NewKinesisDispatcher(config *KinesisDispatcherConfig) *KinesisDispatcher {
	sess := session.Must(session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String(config.Region),
	}))

	cfg := &aws.Config{}

	if config.RoleArn != "" {
		cfg = &aws.Config{
			Credentials: stscreds.NewCredentials(sess, config.RoleArn),
		}
	}

	client := kinesis.New(sess, cfg)
	xray.AWS(client.Client)

	return &KinesisDispatcher{
		StreamName: config.StreamName,
		Client:     client,
	}
}

// HandleContext handles event
func (d *KinesisDispatcher) HandleContext(ctx context.Context, args *EventArgs) error {
	logger := log.WithFields(
		log.Map{
			"event_id":     args.Event.ID,
			"event_name":   args.Event.Name,
			"event_sender": args.Event.Sender,
			"event_source": args.Event.Source,
			"event_target": d.StreamName,
		},
	)

	logger.Info("marshaling event")
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}

	entry := &kinesis.PutRecordInput{
		StreamName:   aws.String(d.StreamName),
		PartitionKey: aws.String(args.Event.ID),
		Data:         data,
	}

	logger.Info("dispatching event to kinesis")
	_, err = d.Client.PutRecordWithContext(ctx, entry)
	return err
}

type (
	// KinesisRecord represents a strem record
	KinesisRecord = consumer.Record

	// KinesisScanFunc is a function executed on kinesis input stream
	KinesisScanFunc = consumer.ScanFunc

	// KinesisCollectorOption represents a collector options
	KinesisCollectorOption = consumer.Option
)

//go:generate counterfeiter -fake-name KinesisScanner -o ./fake/kinesis_scanner.go . KinesisScanner

// KinesisScanner scans kinesis stream
type KinesisScanner interface {
	Scan(ctx context.Context, fn KinesisScanFunc) error
}

// KinesisCollectorConfig represents the kinesis collector config
type KinesisCollectorConfig struct {
	RoleArn      string
	Region       string
	StreamName   string
	Options      []KinesisCollectorOption
	EventHandler EventHandler
}

// KinesisCollector handles kinesis stream
type KinesisCollector struct {
	Scanner      KinesisScanner
	EventHandler EventHandler
	Cancel       context.CancelFunc
}

// NewKinesisCollector creates a new collector to kinesis
func NewKinesisCollector(config *KinesisCollectorConfig) *KinesisCollector {
	sess := session.Must(session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String(config.Region),
	}))

	cfg := &aws.Config{}

	if config.RoleArn != "" {
		cfg = &aws.Config{
			Credentials: stscreds.NewCredentials(sess, config.RoleArn),
		}
	}

	// setup the client
	client := kinesis.New(sess, cfg)

	// Set the options
	options := []KinesisCollectorOption{}
	options = append(options, consumer.WithClient(client))
	options = append(options, config.Options...)

	scanner, err := consumer.New(config.StreamName, options...)
	if err != nil {
		panic(err)
	}

	return &KinesisCollector{
		Scanner:      scanner,
		EventHandler: config.EventHandler,
	}
}

// CollectContextAsync handles the kinesis stream asyncronosly
func (h *KinesisCollector) CollectContextAsync(ctx context.Context) error {
	logger := log.GetContext(ctx)
	ctx, h.Cancel = context.WithCancel(ctx)

	go func() {
		logger.Info("start collecting kinesis stream asyncronously")

		if err := h.CollectContext(ctx); err != nil {
			logger.WithError(err).Error("collecting kinesis stream failed")
		}
	}()

	return nil
}

// CollectContext handles the kinesis stream
func (h *KinesisCollector) CollectContext(ctx context.Context) error {
	return h.Scanner.Scan(ctx, func(record *KinesisRecord) error {
		logger := log.GetContext(ctx).WithFields(
			log.Map{
				"kinesis_partition_key":   aws.StringValue(record.PartitionKey),
				"kinesis_sequence_number": aws.StringValue(record.SequenceNumber),
			},
		)

		args := &EventArgs{}

		logger.Info("unmarshaling event")
		if err := json.Unmarshal(record.Data, args); err != nil {
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
		if err := h.EventHandler.HandleContext(ctx, args); err != nil {
			logger.WithError(err).Error("failed to handle event")
			return err
		}

		return nil
	})
}
