package results

type UsersAggregationResult struct {
	CountUserDevices int64 `bson:"count_user_devices"`
}

func (result *UsersAggregationResult) GetCountUserDevices() int64 {
	return result.CountUserDevices
}
