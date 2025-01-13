package test

// import (
// 	"context"
// 	"duolingo/lib/batch-manager/driver/redis"
// 	"log"

// 	// "log"
// 	// "math/rand"
// 	// "sync"
// 	// "time"
// )

// func TestBatchManager() {
// 	manager := redismanager.GetBatchManager(
// 		context.Background(), 
// 		"builder", 
// 		0,
// 		26, 
// 		5,
// 	)

// 	manager.SetConnection("localhost", "6379")

// 	manager.Reset()

// 	batch, _ := manager.Next()

// 	manager.Progress(batch.Id, 3)

// 	log.Println("Done")

	// var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	// 	defer log.Println("Worker #1 terminated")
	// 	defer wg.Done()
	// 	for {
	// 		batch, err := manager.Next()
	// 		if err != nil {
	// 			log.Println(err)
	// 			break
	// 		}
	// 		if batch == nil {
	// 			break
	// 		}
	// 		log.Println("Worker #1 receive batch: " + batch.Id)
	// 		time.Sleep(time.Second / 2)

    // 		randomBool := rand.Intn(2) == 1
	// 		if randomBool {
	// 			log.Println("Worker #1 rollbacked batch: " + batch.Id)
	// 			manager.RollBack(batch.Id)
	// 		} else {
	// 			log.Println("Worker #1 handle batch: " + batch.Id)
	// 		}


	// 	}
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer log.Println("Worker #2 terminated")
	// 	defer wg.Done()
	// 	for {
	// 		batch, err := manager.Next()
	// 		if err != nil {
	// 			log.Println(err)
	// 			break
	// 		}
	// 		if batch == nil {
	// 			break
	// 		}
	// 		log.Println("Worker #2 handle batch: " + batch.Id)
	// 		time.Sleep(time.Second)
	// 	}
	// }()
	
	// wg.Wait()
// }