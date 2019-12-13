package theta_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phogolabs/theta"
	"github.com/phogolabs/theta/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KinesisHandler", func() {
	var (
		reactor *theta.KinesisHandler
		event   *theta.EventArgs
		handler *fake.EventHandler
	)

	BeforeEach(func() {
		event = &theta.EventArgs{
			Event: &theta.Event{
				ID:        "event-001",
				Source:    "twilio",
				Sender:    "members-api",
				Name:      "member_created",
				Timestamp: time.Now(),
			},
			Meta: theta.Metadata{},
			Body: []byte("{}"),
		}

		handler = &fake.EventHandler{}
		reactor = &theta.KinesisHandler{
			EventHandler: handler,
		}
	})

	NewKinesisEvent := func(data interface{}) events.KinesisEvent {
		buffer := &bytes.Buffer{}

		Expect(json.NewEncoder(buffer).Encode(data)).To(Succeed())

		return events.KinesisEvent{
			Records: []events.KinesisEventRecord{{
				Kinesis: events.KinesisRecord{
					Data: buffer.Bytes(),
				},
			}},
		}
	}

	It("reacts on event successfully", func() {
		kevent := NewKinesisEvent(event)
		Expect(reactor.HandleContext(context.TODO(), kevent)).To(Succeed())

		Expect(handler.HandleContextCallCount()).To(Equal(1))
		_, args := handler.HandleContextArgsForCall(0)
		Expect(args.Event.ID).To(Equal(event.Event.ID))
	})

	Context("when the event cannot be unmarshalled", func() {
		It("returns an error", func() {
			kevent := NewKinesisEvent("")
			Expect(reactor.HandleContext(context.TODO(), kevent)).To(MatchError("json: cannot unmarshal string into Go value of type theta.EventArgs"))
		})
	})

	Context("when the event validation fails", func() {
		BeforeEach(func() {
			event.Event.Name = ""
		})

		It("returns an error", func() {
			kevent := NewKinesisEvent(event)
			Expect(reactor.HandleContext(context.TODO(), kevent)).To(HaveOccurred())
		})
	})

	Context("when the event handler fails", func() {
		BeforeEach(func() {
			handler.HandleContextReturns(fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			kevent := NewKinesisEvent(event)
			Expect(reactor.HandleContext(context.TODO(), kevent)).To(MatchError("oh no"))
		})
	})
})

var _ = Describe("KinesisDispatcher", func() {
	var (
		eventArgs *theta.EventArgs
		handler   *theta.KinesisDispatcher
		client    *fake.KinesisClient
	)

	BeforeEach(func() {
		eventArgs = &theta.EventArgs{
			Event: &theta.Event{ID: "event-001"},
		}

		client = &fake.KinesisClient{}
		handler = &theta.KinesisDispatcher{
			StreamName: "events",
			Client:     client,
		}
	})

	It("streams the event successfully", func() {
		Expect(handler.HandleContext(context.TODO(), eventArgs)).To(Succeed())
		Expect(client.PutRecordWithContextCallCount()).To(Equal(1))

		_, input, _ := client.PutRecordWithContextArgsForCall(0)
		Expect(input.StreamName).To(Equal(&handler.StreamName))
		Expect(input.PartitionKey).To(Equal(&eventArgs.Event.ID))

		outboundEventArgs := &theta.EventArgs{}
		Expect(json.Unmarshal(input.Data, outboundEventArgs)).To(Succeed())
		Expect(outboundEventArgs.Event.ID).To(Equal(eventArgs.Event.ID))
	})

	Context("when the client fails", func() {
		BeforeEach(func() {
			client.PutRecordWithContextReturns(nil, fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			Expect(handler.HandleContext(context.TODO(), eventArgs)).To(MatchError("oh no"))
		})
	})
})
