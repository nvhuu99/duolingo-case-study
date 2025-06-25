package workloads

import (
	"duolingo/models"
	"encoding/json"
	"errors"
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

func (job *TokenBatchJob) Validate() error {
	if job.Message == nil || job.JobId == "" {
		return errors.New("message or job id missing")
	}
	return nil
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
