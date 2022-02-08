package theta

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phogolabs/log"
)

// DynamoHandler creates commands and dispatch them
type DynamoHandler struct {
	EventEncoder EventEncoder
	EventHandler EventHandler
}

// HandleContext dispatches the command to the executor
func (r *DynamoHandler) HandleContext(ctx context.Context, input events.DynamoDBEvent) error {
	for index, record := range input.Records {
		logger := log.WithFields(
			log.Map{
				"dynamo_event":        record.EventID,
				"dynamo_event_name":   record.EventName,
				"dynamo_source":       record.EventSource,
				"dynamo_source_arn":   record.EventSourceArn,
				"dynamo_record_index": index,
			},
		)

		logger.Info("convert a record as map")
		// encode the map
		obj := r.create(record.Change.NewImage)

		logger.Info("marshaling attributes")
		// encode the event
		body, err := r.encode(obj)
		if err != nil {
			logger.WithError(err).Error("failed to marshal attributes")
			return err
		}

		args := &EventArgs{
			Event: &Event{
				ID:        record.EventID,
				Name:      r.event(record),
				Source:    record.EventSource,
				Sender:    "dynamodb",
				Timestamp: time.Now(),
			},
			Body: body,
		}

		ctx = log.SetContext(ctx, logger)

		logger.Info("handling event")
		if err := r.EventHandler.HandleContext(ctx, args); err != nil {
			logger.WithError(err).Error("failed to handle event")
			return err
		}
	}

	return nil
}

func (r *DynamoHandler) event(record events.DynamoDBEventRecord) string {
	var (
		parts = strings.Split(record.EventSourceArn, ":")
		path  = strings.Split(parts[5], "/")
		table = strings.Replace(path[1], "-", "_", -1)
		event = fmt.Sprintf("%s_%s", record.EventName, table)
	)

	return strings.ToLower(event)
}

func (r *DynamoHandler) create(input map[string]events.DynamoDBAttributeValue) map[string]interface{} {
	attributes := make(map[string]interface{})

	for k, v := range input {
		switch v.DataType() {
		case events.DataTypeBinary:
			attributes[k] = v.Binary()
		case events.DataTypeBoolean:
			attributes[k] = v.Boolean()
		case events.DataTypeBinarySet:
			attributes[k] = v.BinarySet()
		case events.DataTypeList:
			attributes[k] = v.List()
		case events.DataTypeMap:
			attributes[k] = r.create(v.Map())
		case events.DataTypeNumber:
			if value, err := v.Integer(); err == nil {
				attributes[k] = value
			} else if value, err := v.Float(); err == nil {
				attributes[k] = value
			} else {
				attributes[k] = v.Number()
			}
		case events.DataTypeNumberSet:
			attributes[k] = v.NumberSet()
		case events.DataTypeNull:
			//TODO: skip
		case events.DataTypeString:
			attributes[k] = v.String()
		case events.DataTypeStringSet:
			attributes[k] = v.StringSet()
		}
	}

	return attributes
}

func (r *DynamoHandler) encode(obj interface{}) ([]byte, error) {
	if encoder := r.EventEncoder; encoder == nil {
		return json.Marshal(obj)
	}

	return r.EventEncoder.Encode(obj)
}
