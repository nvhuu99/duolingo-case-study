package test

// import (
// 	"context"
// 	"duolingo/libraries/events"
// 	"testing"
// 	"time"
// )

// func TestEventManager(t *testing.T) {
// 	events.Init(context.Background(), 2*time.Second)

// 	manager := events.GetManager()

// 	manager.AddSubsriber(&events.SubscriberImp{Name: "sub1"})

// 	ctxA, A := events.New(context.Background(), "A")

// 	ctxB, B := events.New(ctxA, "B")

// 	ctxC, C := events.New(ctxA, "C")

// 	_, D := events.New(ctxB, "D")

// 	ctxC1, cancelCtxC1 := context.WithCancel(ctxC)
// 	ctxE, _ := events.New(ctxC1, "E")

// 	_, F := events.New(ctxC, "F")

// 	_, G := events.New(ctxE, "G")
// 	_, H := events.New(ctxE, "H")

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(A)

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(B)

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(C)

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(D)

// 	time.Sleep(100 * time.Millisecond)
// 	// events.Interupt(E)
// 	cancelCtxC1()

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(F)

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(G)

// 	time.Sleep(100 * time.Millisecond)
// 	events.End(H)

// 	time.Sleep(15 * time.Second)
// }
