package main

import (
	"context"
	"duolingo/constants"
	"duolingo/libraries/pub_sub"
	container "duolingo/libraries/service_container"
	"duolingo/models"
	"duolingo/services/noti_builder/server/tasks"
	"duolingo/services/noti_builder/server/workloads"
	"sync"
)

func main() {
	inputTopic := constants.TopicMessageInputs
	jobTopic := constants.TopicNotiBuilderJobs
	inputConsumer := container.MustResolve[pub_sub.Subscriber]()
	jobConsumer := container.MustResolve[pub_sub.Subscriber]()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 2)
	go func() {
		if err := <-errChan; err != nil {
			cancel()
		}
	}()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := inputConsumer.Consuming(ctx, inputTopic, func(str string) pub_sub.ConsumeAction {
			return acceptOrReject(
				tasks.HandleMessageInput(models.MessageInputDecode([]byte(str))),
			)
		})
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := jobConsumer.Consuming(ctx, jobTopic, func(str string) pub_sub.ConsumeAction {
			return acceptOrReject(
				tasks.CreatePushNotiMessageBatches(ctx, workloads.JobDecode([]byte(str))),
			)
		})
		if err != nil {
			errChan <- err
		}
	}()
	wg.Wait()
}

func acceptOrReject(err error) pub_sub.ConsumeAction {
	if err != nil {
		return pub_sub.ActionAccept
	}
	return pub_sub.ActionReject
}
