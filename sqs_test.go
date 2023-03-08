package theta_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/phogolabs/theta"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/phogolabs/theta/fake"
)

var _ = Describe("SQSHandler", func() {
	var (
		reactor *theta.SQSHandler
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
		reactor = &theta.SQSHandler{
			EventHandler: handler,
		}
	})

	NewSQSEvent := func(data interface{}) events.SQSEvent {
		buffer := &bytes.Buffer{}

		writer := base64.NewEncoder(
			base64.StdEncoding,
			buffer,
		)

		Expect(json.NewEncoder(writer).Encode(data)).To(Succeed())
		Expect(writer.Close()).To(Succeed())

		return events.SQSEvent{
			Records: []events.SQSMessage{
				{Body: buffer.String()},
			},
		}
	}

	It("reacts on event successfully", func() {
		sqsevent := NewSQSEvent(event)
		Expect(reactor.HandleContext(context.TODO(), sqsevent)).To(Succeed())

		Expect(handler.HandleContextCallCount()).To(Equal(1))
		_, args := handler.HandleContextArgsForCall(0)
		Expect(args.Event.ID).To(Equal(event.Event.ID))
	})

	Context("when the event cannot be unmarshalled", func() {
		It("returns an error", func() {
			sqsevent := NewSQSEvent("")
			Expect(reactor.HandleContext(context.TODO(), sqsevent)).To(MatchError("json: cannot unmarshal string into Go value of type theta.EventArgs"))
		})
	})

	Context("when the event validation fails", func() {
		BeforeEach(func() {
			event.Event.Name = ""
		})

		It("returns an error", func() {
			sqsevent := NewSQSEvent(event)
			Expect(reactor.HandleContext(context.TODO(), sqsevent)).To(HaveOccurred())
		})
	})

	Context("when the event handler fails", func() {
		BeforeEach(func() {
			handler.HandleContextReturns(fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			sqsevent := NewSQSEvent(event)
			Expect(reactor.HandleContext(context.TODO(), sqsevent)).To(MatchError("oh no"))
		})
	})
})

// var _ = Describe("SQSCommandDispatcher", func() {
// 	var (
// 		command    *domain.CommandArgs
// 		dispatcher *domain.SQSCommandDispatcher
// 		client     *FakeSQSClient
// 	)

// 	BeforeEach(func() {
// 		client = &FakeSQSClient{}

// 		dispatcher = &domain.SQSCommandDispatcher{
// 			QueueURL: "http://example.com",
// 			Client:   client,
// 		}

// 		command = &domain.CommandArgs{
// 			Command: &domain.Command{
// 				ID:        "cmd-001",
// 				Name:      "send",
// 				Sender:    "notification-api",
// 				Source:    "test",
// 				Timestamp: time.Now(),
// 			},
// 			Param: []byte("{}"),
// 		}
// 	})

// 	It("dispatches the command successfully", func() {
// 		Expect(dispatcher.ExecuteContext(context.TODO(), command)).To(Succeed())

// 		Expect(client.SendMessageWithContextCallCount()).To(Equal(1))
// 		ctx, input, _ := client.SendMessageWithContextArgsForCall(0)
// 		Expect(ctx).To(Equal(context.TODO()))
// 		Expect(input.QueueUrl).To(Equal(aws.String("http://example.com")))

// 		reader := base64.NewDecoder(
// 			base64.StdEncoding,
// 			bytes.NewBufferString(aws.StringValue(input.MessageBody)),
// 		)

// 		executed := &domain.CommandArgs{}
// 		Expect(json.NewDecoder(reader).Decode(executed)).To(Succeed())
// 		Expect(executed.Command.Name).To(Equal(command.Command.Name))
// 	})

// 	Context("when the command is not valid", func() {
// 		BeforeEach(func() {
// 			command.Command.Name = ""
// 		})

// 		It("returns an error", func() {
// 			err := dispatcher.ExecuteContext(context.TODO(), command)
// 			Expect(err).To(HaveOccurred())
// 		})
// 	})

// 	Context("when the client fails", func() {
// 		BeforeEach(func() {
// 			client.SendMessageWithContextReturns(nil, fmt.Errorf("oh no"))
// 		})

// 		It("returns an error", func() {
// 			Expect(dispatcher.ExecuteContext(context.TODO(), command)).To(MatchError("oh no"))
// 		})
// 	})
// })

var _ = Describe("SQSDispatcher", func() {
	var (
		eventArgs *theta.EventArgs
		handler   *theta.SQSDispatcher
		client    *FakeSQSClient
	)

	BeforeEach(func() {
		eventArgs = &theta.EventArgs{
			Event: &theta.Event{ID: "event-001"},
		}

		client = &FakeSQSClient{}
		handler = &theta.SQSDispatcher{
			QueueURL: "http://example.com",
			Client:   client,
		}
	})

	It("streams the event successfully", func() {
		Expect(handler.HandleContext(context.TODO(), eventArgs)).To(Succeed())

		Expect(client.SendMessageWithContextCallCount()).To(Equal(1))
		ctx, input, _ := client.SendMessageWithContextArgsForCall(0)
		Expect(ctx).To(Equal(context.TODO()))
		Expect(input.QueueUrl).To(Equal(aws.String("http://example.com")))

		reader := base64.NewDecoder(
			base64.StdEncoding,
			bytes.NewBufferString(aws.StringValue(input.MessageBody)),
		)

		outboundEventArgs := &theta.EventArgs{}
		Expect(json.NewDecoder(reader).Decode(outboundEventArgs)).To(Succeed())
		Expect(outboundEventArgs.Event.ID).To(Equal(eventArgs.Event.ID))
	})

	Context("when the client fails", func() {
		BeforeEach(func() {
			client.SendMessageWithContextReturns(nil, fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			Expect(handler.HandleContext(context.TODO(), eventArgs)).To(MatchError("oh no"))
		})
	})
})
