package theta

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phogolabs/log"
)

// CloudWatchHandler handles cloud watch cron jobs
type CloudWatchHandler struct {
	EventHandler EventHandler
}

// HandleContext dispatches the command to the executor
func (h *CloudWatchHandler) HandleContext(ctx context.Context, input events.CloudWatchEvent) error {
	logger := log.WithFields(
		log.Map{
			"cloud_watch_event_id":     input.ID,
			"cloud_watch_event_source": input.Source,
			"cloud_watch_detail_type":  input.DetailType,
		},
	)

	args := &EventArgs{
		Event: &Event{
			ID:        input.ID,
			Name:      input.DetailType,
			Source:    input.Source,
			Timestamp: input.Time,
			Sender:    "cloudwatch",
		},
		Body: input.Detail,
	}

	ctx = log.SetContext(ctx, logger)

	logger.Info("handling event")
	if err := h.EventHandler.HandleContext(ctx, args); err != nil {
		logger.WithError(err).Error("failed to handle event")
		return err
	}

	return nil
}
