package workloads

import (
	"duolingo/models"
	"encoding/json"
)

type TokenBatchJob struct {
	JobId   string               `json:"job_id"`
	Message *models.MessageInput `json:"message"`
}

func NewTokenBatchJob(jobId string, message *models.MessageInput) *TokenBatchJob {
	return &TokenBatchJob{
		JobId:   jobId,
		Message: message,
	}
}

func (job *TokenBatchJob) Encode() []byte {
	marshalled, err := json.Marshal(job)
	if err != nil {
		panic(err)
	}
	return marshalled
}

func JobDecode(data []byte) *TokenBatchJob {
	job := new(TokenBatchJob)
	err := json.Unmarshal(data, job)
	if err != nil {
		panic(err)
	}
	return job
}
