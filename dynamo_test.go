package theta_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phogolabs/theta"
	"github.com/phogolabs/theta/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DynamoHandler", func() {
	var (
		reactor *theta.DynamoHandler
		handler *fake.EventHandler
	)

	BeforeEach(func() {
		handler = &fake.EventHandler{}
		reactor = &theta.DynamoHandler{
			EventHandler: handler,
		}
	})

	NewDynamoEvent := func(data map[string]events.DynamoDBAttributeValue) events.DynamoDBEvent {
		return events.DynamoDBEvent{
			Records: []events.DynamoDBEventRecord{
				{
					EventName:      "INSERT",
					EventSourceArn: "arn:aws:dynamodb:us-east-1:655141976367:table/notification-messages/stream/2019-07-11T10:10:31.424",
					Change:         events.DynamoDBStreamRecord{NewImage: data},
				},
			},
		}
	}

	It("reacts on event successfully", func() {
		attr := make(map[string]events.DynamoDBAttributeValue)
		attr["id"] = events.NewStringAttribute("123")

		kevent := NewDynamoEvent(attr)
		Expect(reactor.HandleContext(context.TODO(), kevent)).To(Succeed())

		Expect(handler.HandleContextCallCount()).To(Equal(1))
		_, args := handler.HandleContextArgsForCall(0)
		Expect(args.Event.Name).To(Equal("insert_notification_messages"))

		type Row struct {
			ID string `json:"id"`
		}

		row := &Row{}

		Expect(json.Unmarshal(args.Body, row)).To(Succeed())
		Expect(row.ID).To(Equal("123"))
	})

	Context("when the event handler fails", func() {
		BeforeEach(func() {
			handler.HandleContextReturns(fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			attr := make(map[string]events.DynamoDBAttributeValue)

			kevent := NewDynamoEvent(attr)
			Expect(reactor.HandleContext(context.TODO(), kevent)).To(MatchError("oh no"))
		})
	})
})
