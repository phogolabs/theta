package theta_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/phogolabs/theta"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/phogolabs/theta/fake"
)

var _ = Describe("KinesisHandler", func() {
	var (
		reactor *theta.KinesisHandler
		event   *theta.EventArgs
		handler *FakeEventHandler
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

		handler = &FakeEventHandler{}
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
		client    *FakeKinesisClient
	)

	BeforeEach(func() {
		eventArgs = &theta.EventArgs{
			Event: &theta.Event{ID: "event-001"},
		}

		client = &FakeKinesisClient{}
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

var _ = Describe("KinesisCollector", func() {
	var (
		collector *theta.KinesisCollector
		handler   *FakeEventHandler
		event     *theta.EventArgs
		scanner   *FakeKinesisScanner
	)

	NewKinesisRecord := func(data interface{}) theta.KinesisRecord {
		buffer := &bytes.Buffer{}
		Expect(json.NewEncoder(buffer).Encode(data)).To(Succeed())

		return theta.KinesisRecord{
			Record: &kinesis.Record{
				Data: buffer.Bytes(),
			},
		}
	}

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

		handler = &FakeEventHandler{}

		scanner = &FakeKinesisScanner{}
		scanner.ScanStub = func(ctx context.Context, fn theta.KinesisScanFunc) error {
			row := NewKinesisRecord(event)
			return fn(&row)
		}

		collector = &theta.KinesisCollector{
			Scanner:      scanner,
			EventHandler: handler,
		}
	})

	It("handles a record successfully", func() {
		Expect(collector.CollectContext(context.TODO())).To(Succeed())
		Expect(scanner.ScanCallCount()).To(Equal(1))
		Expect(handler.HandleContextCallCount()).To(Equal(1))

		_, args := handler.HandleContextArgsForCall(0)
		Expect(args.Event.ID).To(Equal("event-001"))
	})

	Context("when the event unmarshaling fails", func() {
		BeforeEach(func() {
			scanner.ScanStub = func(ctx context.Context, fn theta.KinesisScanFunc) error {
				row := &theta.KinesisRecord{
					Record: &kinesis.Record{
						Data: []byte(`{"member_id":`),
					},
				}

				return fn(row)
			}
		})

		It("returns an error", func() {
			Expect(collector.CollectContext(context.TODO())).To(MatchError("unexpected EOF"))
		})
	})

	Context("when the event validation fails", func() {
		BeforeEach(func() {
			scanner.ScanStub = func(ctx context.Context, fn theta.KinesisScanFunc) error {
				event.Event.ID = ""
				row := NewKinesisRecord(event)
				return fn(&row)
			}
		})

		It("returns an error", func() {
			Expect(collector.CollectContext(context.TODO())).To(MatchError("Key: 'EventArgs.Event.ID' Error:Field validation for 'ID' failed on the 'required' tag"))
		})
	})

	Context("when the handler fails", func() {
		BeforeEach(func() {
			handler.HandleContextReturns(fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			Expect(collector.CollectContext(context.TODO())).To(MatchError("oh no"))
		})
	})
})
