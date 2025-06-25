package server

import (
	"context"
	"duolingo/libraries/pub_sub"
	"duolingo/models"
	wrkl "duolingo/services/noti_builder/server/workloads"
	"sync"
)

type NotiBuilder struct {
	inputSubscriber  pub_sub.Subscriber
	notiPublisher    pub_sub.Publisher
	tokenDistributor *wrkl.TokenBatchDistributor
}

func NewNotiBuilder(
	inputSubscriber pub_sub.Subscriber,
	notiPublisher pub_sub.Publisher,
	tokenDistributor *wrkl.TokenBatchDistributor,
) *NotiBuilder {
	return &NotiBuilder{
		inputSubscriber:  inputSubscriber,
		notiPublisher:    notiPublisher,
		tokenDistributor: tokenDistributor,
	}
}

func (b *NotiBuilder) Start(buildCtx context.Context) error {
	var buildErr error
	ctx, cancel := context.WithCancel(buildCtx)
	defer cancel()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer cancel()
		buildErr = b.inputSubscriber.ConsumingMainTopic(ctx, func(str string) pub_sub.ConsumeAction {
			return b.acceptOrReject(
				b.tokenDistributor.CreateBatchJob(models.MessageInputDecode([]byte(str))))
		})
	}()
	go func() {
		defer wg.Done()
		defer cancel()
		buildErr = b.tokenDistributor.ConsumingTokenBatches(ctx, func(
			input *models.MessageInput,
			devices []*models.UserDevice,
		) error {
			return b.notiPublisher.NotifyMainTopic(
				string(models.NewPushNotiMessage(input, devices).Encode()))
		})
	}()
	wg.Wait()

	return buildErr
}

func (b *NotiBuilder) acceptOrReject(err error) pub_sub.ConsumeAction {
	if err != nil {
		return pub_sub.ActionAccept
	}
	return pub_sub.ActionReject
}
