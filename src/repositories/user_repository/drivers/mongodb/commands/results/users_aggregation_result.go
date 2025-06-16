package results

type UsersAggregationResult struct {
	CountUserDevices uint64 `bson:"count_user_devices"`
}

func (result *UsersAggregationResult) GetCountUserDevices() uint64 {
	return result.CountUserDevices
}
