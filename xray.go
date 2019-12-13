package theta

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/phogolabs/log"
)

var _ EventHandler = &XrayEventHandler{}

// XrayEventHandler is a handler that add XRay segment
type XrayEventHandler struct {
	Segment      string
	Metadata     map[string]interface{}
	EventHandler EventHandler
}

// HandleContext handles the event
func (h *XrayEventHandler) HandleContext(ctx context.Context, args *EventArgs) error {
	var (
		logger  = log.GetContext(ctx)
		parent  = xray.GetSegment(ctx)
		segment *xray.Segment
	)

	if parent == nil {
		ctx, segment = xray.BeginSegment(ctx, h.Segment)
	} else {
		ctx, segment = xray.BeginSubsegment(ctx, h.Segment)
	}

	annotations := map[string]string{
		"event_id":     args.Event.ID,
		"event_name":   args.Event.Name,
		"event_source": args.Event.Source,
		"event_sender": args.Event.Sender,
	}

	for k, v := range annotations {
		if err := segment.AddAnnotation(k, v); err != nil {
			logger.WithError(err).Errorf("failed to add annotation for key %s with value %s", k, v)
		}
	}

	if h.Metadata != nil {
		for k, v := range h.Metadata {
			if err := segment.AddMetadata(k, v); err != nil {
				logger.WithError(err).Errorf("failed to add metadata for key %s with value %s", k, v)
			}
		}
	}

	err := h.EventHandler.HandleContext(ctx, args)
	segment.Close(err)

	return err
}
