package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"duolingo/libraries/buffer"
)

type LogWriter interface {
	Write(log *Log)
}

/* Console Writer */

type ConsoleWriter struct {
	formatter LogFormatter
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

func (writer *ConsoleWriter) WithFormatter(formatter LogFormatter) *ConsoleWriter {
	writer.formatter = formatter
	return writer
}

func (writer *ConsoleWriter) Write(log *Log) {
	formatted, err := writer.formatter.Format(log)
	if err != nil {
		fmt.Println("level: error - ns: telementry.otel_wrapper.log.console_writer - message: log format failed")
	} else {
		fmt.Println(formatted)
	}
}

/* Grafana writerWriter */

type LokiWriter struct {
	serviceName  string
	lokiEndpoint string

	formatter LogFormatter
	buffer    *buffer.BufferGroup[LogLevel, *Log]
}

func NewLokiWriter(
	ctx context.Context,
	serviceName string,
	lokiEndpoint string,
	limit int,
	interval time.Duration,
) *LokiWriter {
	writer := &LokiWriter{
		serviceName:  serviceName,
		lokiEndpoint: lokiEndpoint,
		buffer:       buffer.NewBufferGroup[LogLevel, *Log](),
	}

	writer.buffer.
		SetLimit(limit).
		SetInterval(interval).
		SetConsumeFunc(false, writer.flush).
		DeclareGroup(ctx, LevelError).
		DeclareGroup(ctx, LevelInfo).
		DeclareGroup(ctx, LevelDebug)

	return writer
}

func (writer *LokiWriter) WithFormatter(formatter LogFormatter) *LokiWriter {
	writer.formatter = formatter
	return writer
}

func (writer *LokiWriter) Write(log *Log) {
	writer.buffer.Write(log.Level, log)
}

func (writer *LokiWriter) flush(ctx context.Context, level LogLevel, logs []*Log) {

	if len(logs) == 0 {
		return
	}

	labels := map[string]string{
		"level":        logLevelAsString[level],
		"service_name": writer.serviceName,
	}
	entries := make([][]string, len(logs))
	for i, log := range logs {
		formatted, _ := writer.formatter.Format(log)
		entries[i] = []string{
			fmt.Sprintf("%d", log.Timestamp.UnixNano()),
			formatted,
		}
	}

	requestBody, _ := json.Marshal(LokiPushRequest{
		Streams: []LokiLogEntry{{
			Stream: labels,
			Values: entries,
		}},
	})
	req, err := http.NewRequest("POST", writer.lokiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		panic("unable to construct Loki push api request, err: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		fmt.Println("level: error - ns: telementry.otel_wrapper.log.loki_writer - message: failed to flush log - err: " + err.Error())
	}
}

type LokiLogEntry struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type LokiPushRequest struct {
	Streams []LokiLogEntry `json:"streams"`
}
