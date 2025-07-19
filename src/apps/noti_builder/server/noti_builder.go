package server

import (
	"context"
	"log"
	"sync"

	wrkl "duolingo/apps/noti_builder/server/workloads"
	container "duolingo/libraries/dependencies_container"
	ps "duolingo/libraries/message_queue/pub_sub"
	tq "duolingo/libraries/message_queue/task_queue"
	"duolingo/models"
)

type NotiBuilder struct {
	inputSubscriber      ps.Subscriber
	pushNotiTaskProducer tq.TaskProducer
	tokenDistributor     *wrkl.TokenBatchDistributor
	errChan              chan error
}

func NewNotiBuilder() *NotiBuilder {
	return &NotiBuilder{
		inputSubscriber:      container.MustResolveAlias[ps.Subscriber]("message_input_subscriber"),
		pushNotiTaskProducer: container.MustResolveAlias[tq.TaskProducer]("push_notifications_producer"),
		tokenDistributor:     wrkl.NewTokenBatchDistributor(),
		errChan:              make(chan error, 100),
	}
}

func (b *NotiBuilder) Start(buildCtx context.Context) {
	ctx, cancel := context.WithCancel(buildCtx)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go b.handleErrChannel(wg, ctx)

	go func() {
		defer wg.Done()
		defer cancel()
		if err := b.inputSubscriber.ListeningMainTopic(ctx, b.createBatchJob); err != nil {
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

func (b *NotiBuilder) handleErrChannel(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-b.errChan:
			if err != nil {
				log.Println("err", err)
			}
		}
	}
}

func (b *NotiBuilder) createBatchJob(ctx context.Context, serialized string) {
	inp := models.MessageInputDecode([]byte(serialized))
	if err := b.tokenDistributor.CreateBatchJob(inp); err != nil {
		b.errChan <- err
	}
	log.Println("job queued:", serialized)
}

func (b *NotiBuilder) producePushNotiTask(
	input *models.MessageInput,
	devices []*models.UserDevice,
) {
	serialized := string(models.NewPushNotiMessage(input, devices).Encode())
	if err := b.pushNotiTaskProducer.Push(serialized); err != nil {
		b.errChan <- err
	}
	log.Println("task pushed:", serialized)
}
