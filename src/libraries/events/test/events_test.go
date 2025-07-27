package test

import (
	"context"
	"duolingo/libraries/events"
	"testing"
	"time"
)

func TestEvents(t *testing.T) {
	manager := events.GetManager()

	manager.AddSubsriber(&events.SubscriberImp{ Name: "sub1" })
	// manager.AddSubsriber(&events.SubscriberImp{ Name: "sub2" })

	ctxA, A := events.New(context.Background(), "A")
	
	ctxB, B := events.New(ctxA, "B")
	ctxC, C := events.New(ctxA, "C")

	_, D := events.New(ctxB, "D")
	ctxE, E := events.New(ctxC, "E")
	_, F := events.New(ctxC, "F")

	_, G := events.New(ctxE, "G")
	_, H := events.New(ctxE, "H")

	events.End(A)
	events.End(B)
	events.End(C)
	events.End(D)
	events.End(E)
	events.End(F)
	events.End(G)
	events.End(H)

	time.Sleep(15*time.Second)
}