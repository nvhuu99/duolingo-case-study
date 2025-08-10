package server

import (
	"context"
	"log"
	"sync"

	wrkl "duolingo/apps/noti_builder/server/workloads"
	container "duolingo/libraries/dependencies_container"
	events "duolingo/libraries/events/facade"
	ps "duolingo/libraries/message_queue/pub_sub"
	tq "duolingo/libraries/message_queue/task_queue"
	"duolingo/models"
)

type NotiBuilder struct {
	msgInpSubscriber ps.Subscriber
	pushNotiProducer tq.TaskProducer
	tokenDistributor *wrkl.TokenBatchDistributor
}

func NewNotiBuilder() *NotiBuilder {
	return &NotiBuilder{
		msgInpSubscriber: container.MustResolveAlias[ps.Subscriber]("message_input_subscriber"),
		pushNotiProducer: container.MustResolveAlias[tq.TaskProducer]("push_notifications_producer"),
		tokenDistributor: wrkl.NewTokenBatchDistributor(),
	}
}

func (b *NotiBuilder) Start(buildCtx context.Context) {
	log.Println("running noti builder")

	ctx, cancel := context.WithCancel(buildCtx)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer cancel()
		if err := b.msgInpSubscriber.ListeningMainTopic(ctx, b.createBatchJob); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		if err := b.tokenDistributor.ConsumingTokenBatches(ctx, b.producePushNotiTask); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func (b *NotiBuilder) createBatchJob(ctx context.Context, serialized string) error {
	var err error

	evt := events.Start(ctx, "noti_builder.create_batch_job", nil)
	defer events.End(evt, true, err, nil)
	defer log.Println("create batch job, err:", err)

	err = b.tokenDistributor.CreateBatchJob(
		evt.Context(), 
		models.MessageInputDecode([]byte(serialized)),
	)

	return err
}

func (b *NotiBuilder) producePushNotiTask(
	ctx context.Context,
	input *models.MessageInput,
	devices []*models.UserDevice,
) error {
	var err error

	evt := events.Start(ctx, "noti_builder.produce_push_noti_task", nil)
	defer events.End(evt, true, err, nil)
	defer log.Println("produce push noti task, err:", err)

	serialized := string(models.NewPushNotiMessage(input, devices).Encode())
	err = b.pushNotiProducer.Push(evt.Context(), serialized)


	return err
}
