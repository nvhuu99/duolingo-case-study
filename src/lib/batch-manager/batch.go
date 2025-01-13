package batchmanager

type BatchItem struct {
	Id string `json:"id"`
	Start int `json:"start"`
	End int `json:"end"`
	Progress int `json:"progress"`
	HasFailed bool `json:"hasFailed"`
}