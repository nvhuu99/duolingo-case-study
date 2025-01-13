package redismanager

const NilBatch NilBatchError = "batch manager: all batches have been proccessed" 

type NilBatchError string

func (e NilBatchError) Error() string { return string(e) }
