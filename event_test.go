package theta_test

import (
	"context"
	"fmt"
	"time"

	"github.com/phogolabs/theta"
	"github.com/phogolabs/theta/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompositeEventHandler", func() {
	var (
		reactor *theta.CompositeEventHandler
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
		reactor = &theta.CompositeEventHandler{
			handler,
		}
	})

	It("reacts on event successfully", func() {
		Expect(reactor.HandleContext(context.TODO(), event)).To(Succeed())
		Expect(handler.HandleContextCallCount()).To(Equal(1))

		_, args := handler.HandleContextArgsForCall(0)
		Expect(args.Event.ID).To(Equal(event.Event.ID))
	})

	Context("when the event handler fails", func() {
		BeforeEach(func() {
			handler.HandleContextReturns(fmt.Errorf("oh no"))
		})

		It("returns an error", func() {
			Expect(reactor.HandleContext(context.TODO(), event)).To(MatchError("oh no"))
		})
	})
})
