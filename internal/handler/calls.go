package handler

import (
	"context"
	"fmt"
	"github.com/webitel/web-meeting-backend/infra/pubsub"
	"github.com/webitel/web-meeting-backend/internal/model"
	"github.com/webitel/wlog"
)

type CallsHandler struct {
	log    *wlog.Logger
	svc    MeetingService
	pubSub *pubsub.Manager
}

func NewCallsHandler(svc MeetingService, pubSub *pubsub.Manager, l *wlog.Logger) *CallsHandler {
	ch := &CallsHandler{
		svc:    svc,
		pubSub: pubSub,
		log:    l,
	}
	pubSub.AddOnConnect(func(channel *pubsub.Channel) error {
		var err error
		var delivery pubsub.Delivery
		const exchange = "call"
		const queueName = "call-meetings"

		if err = channel.DeclareExchange(pubsub.Exchange{
			Name:    exchange,
			Type:    pubsub.ExchangeTypeTopic,
			Durable: true,
		}); err != nil {
			return err
		}

		if err = channel.DeclareDurableQueue(queueName, pubsub.Headers{
			"x-queue-type": "quorum",
		}); err != nil {
			return err
		}

		if err = channel.BindQueue(queueName, "events.hangup.*.*.*", exchange, pubsub.Headers{
			"x-expires": 5 * 60 * 1000, // 5 minutes
		}); err != nil {
			return err
		}

		delivery, err = channel.ConsumeQueue(queueName, false)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case msg, ok := <-delivery:
					if !ok {
						return
					}

					c, err := model.CallFromJson(msg.Body)
					if err != nil {
						l.Error("failed to parse call", wlog.Err(err))
						msg.Ack(true)
						continue
					}

					if c.Data.MeetingId == nil {
						l.Debug(fmt.Sprintf("skip call [%s] without meeting id", c.Id))
						msg.Ack(true)
						continue
					}

					ctx := context.Background()
					var id string

					id, err = svc.CloseByCall(ctx, *c.Data.MeetingId, c.Id)
					if err != nil {
						l.Error("failed to set call_id", wlog.Err(err))
					}

					l.Debug(fmt.Sprintf("call [%s] finished; meeting_id: %s", c.Id, id))
					msg.Ack(true)
				}
			}
		}()

		return nil
	})

	return ch
}
