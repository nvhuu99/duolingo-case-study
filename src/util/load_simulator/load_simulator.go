package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	campaign := flag.String("campaign", "", "")
	numReq := flag.Int("num", 0, "")
	flag.Parse()

	if *campaign == "" {
		log.Println("params invalid. no request made")
		return
	}

	for range *numReq {
		message := []byte(fmt.Sprintf("{ \"title\": \"%v\", \"content\": \"%v\" }", *campaign, *campaign))
		res, err := http.Post(fmt.Sprintf("http://input-message-api:8001/campaign/%v/message", *campaign), "application/json", bytes.NewBuffer(message))
		if err != nil || res.StatusCode != http.StatusCreated {
			log.Println("abort api error", err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("finished input %v messages for the %v campaign\n", *numReq, *campaign)
}
