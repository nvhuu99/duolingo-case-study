package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	ed "duolingo/event/event_data"
	eh "duolingo/event/event_handler"
	cf "duolingo/lib/config_reader"
	ep "duolingo/lib/event"
	mq "duolingo/lib/message_queue"
	sv "duolingo/lib/service_container"
	wd "duolingo/lib/work_distributor"
	md "duolingo/model"
	db "duolingo/repository/campaign_db"
	"duolingo/service/noti_builder/bootstrap"

	"github.com/google/uuid"
)

var (
	ctx       context.Context
	container *sv.ServiceContainer
	conf      cf.ConfigReader

	repo        *db.UserRepo
	distributor wd.Distributor
	event       *ep.EventPublisher

	consumer           mq.Consumer
	inputMssgPublisher mq.Publisher
	pushNotiPublisher  mq.Publisher
)

func main() {
	bootstrap.Run()

	container = sv.GetContainer()
	ctx = container.Resolve("server.ctx").(context.Context)
	conf = container.Resolve("config").(cf.ConfigReader)
	repo = container.Resolve("repo.campaign_user").(*db.UserRepo)
	distributor = container.Resolve("distributor").(wd.Distributor)
	consumer = container.Resolve("mq.consumer.input_messages").(mq.Consumer)
	event = container.Resolve("event.publisher").(*ep.EventPublisher)
	inputMssgPublisher = container.Resolve("mq.publisher.input_messages").(mq.Publisher)
	pushNotiPublisher = container.Resolve("mq.publisher.push_noti_messages").(mq.Publisher)

	log.Println("builder started")

	consumer.Consume(make(chan bool, 1), func(body []byte) mq.ConsumerAction {
		pushNoti := new(md.PushNotiMessage)
		json.Unmarshal(body, pushNoti)
		if pushNoti.RelayFlag == md.ShouldRelay {
			return relay(pushNoti)
		}
		return build(pushNoti)
	})
}

func relay(pushNoti *md.PushNotiMessage) mq.ConsumerAction {
	relayEvent := &ed.RelayInputMessage{OptId: uuid.NewString(), PushNoti: pushNoti}
	event.Notify(true, eh.RELAY_INP_MESG_BEGIN, relayEvent)
	defer event.Notify(true, eh.RELAY_INP_MESG_END, relayEvent)

	// Register a new workload
	count, err := repo.CountCampaignMsgReceivers(
		pushNoti.InputMessage.Campaign, 
		pushNoti.InputMessage.CreatedAt,
	)
	if err != nil {
		relayEvent.Error = err
		return mq.ConsumerRequeue
	}
	if count == 0 {
		relayEvent.MessageIgnoreReason = "no campaign user found"
		relayEvent.Success = true
		return mq.ConsumerAccept
	}
	err = distributor.RegisterWorkLoad(&wd.Workload{
		Name:       pushNoti.InputMessage.MessageId,
		NumOfUnits: count,
	})
	if err != nil {
		relayEvent.Error = err
		return mq.ConsumerRequeue
	}
	// Build relay the message
	trace := container.Resolve("events.data.sv_opt_trace." + relayEvent.OptId).(*ed.ServiceOperationTrace)
	numOfBuilders := conf.GetInt("noti_builder.server.num_of_builders", 1)
	batch := make([]string, numOfBuilders)
	for i := 0; i < numOfBuilders; i++ {
		pushNoti.Trace = trace.Span
		pushNoti.RelayFlag = md.HasRelayed
		serialized, _ := json.Marshal(pushNoti)
		batch[i] = string(serialized)
	}
	// Start relaying
	i := 0
	for {
		select {
		case <-ctx.Done():
			relayEvent.Error = fmt.Errorf("service terminated")
			return mq.ConsumerRequeue
		default:
		}
		if i == len(batch) {
			relayEvent.RelayedCount = uint8(numOfBuilders)
			relayEvent.Success = true
			return mq.ConsumerAccept
		}
		if err := inputMssgPublisher.Publish(batch[i]); err != nil {
			relayEvent.Error = err
			return mq.ConsumerRequeue
		}
		i++
	}
}

func build(pushNoti *md.PushNotiMessage) mq.ConsumerAction {
	var allSuccess bool
	var err error
	var workload *wd.Workload
	var assignment *wd.Assignment
	var assignments []*wd.Assignment

	buildEvent := &ed.BuildPushNotiMessage{OptId: uuid.NewString(), PushNoti: pushNoti}
	event.Notify(true, eh.BUILD_PUSH_NOTI_MESG_BEGIN, buildEvent)
	defer event.Notify(true, eh.BUILD_PUSH_NOTI_MESG_END, buildEvent)
	defer func() {
		buildEvent.Assignments = assignments
		buildEvent.Workload = workload
		buildEvent.Error = err
		buildEvent.Success = allSuccess
	}()

	workload, err = distributor.SwitchToWorkload(pushNoti.InputMessage.MessageId)
	if err != nil {
		return mq.ConsumerRequeue
	}

	for {
		select {
		case <-ctx.Done():
			return mq.ConsumerRequeue
		default:
		}

		assignment, err = distributor.Next()
		if err != nil {
			// empty assignment error indicates operations success
			if assignment == nil {
				allSuccess = true
				err = nil
				return mq.ConsumerAccept
			}
			return mq.ConsumerRequeue
		}
		assignments = append(assignments, assignment)

		deviceTokens, err := repo.ListCampaignMsgReceiverTokens(
			pushNoti.InputMessage.Campaign,
			pushNoti.InputMessage.CreatedAt,
			&db.QueryOptions{
				Skip:     int64(assignment.Start - 1),
				Limit:    int64(assignment.End - assignment.Start + 1),
			},
		)
		if err != nil {
			return mq.ConsumerRequeue
		}

		trace := container.Resolve("events.data.sv_opt_trace." + buildEvent.OptId).(*ed.ServiceOperationTrace)
		pushNoti.Trace = trace.Span
		pushNoti.DeviceTokens = deviceTokens

		if err = pushNotiPublisher.Publish(pushNoti.Serialize()); err != nil {
			return mq.ConsumerRequeue
		}
		distributor.Progress(pushNoti.InputMessage.MessageId, assignment.End)
		distributor.Commit(assignment.Id)
	}
}
