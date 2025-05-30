package constant

const (
	METRIC_TARGET_REDIS = "redis"
	METRIC_NAME_REDIS_CMD_RATE = "command_rate"
	METRIC_NAME_REDIS_LOCK_WAITED = "lock_waited_ms"
	METRIC_NAME_REDIS_LOCK_HELD = "lock_held_ms"

	METRIC_TARGET_RABBITMQ = "rabbitmq"
	METRIC_NAME_DELIVERED_RATE = "delivered_rate"
	METRIC_NAME_PUBLISHED_RATE = "published_rate"
	METRIC_NAME_PUBLISH_LATENCY = "publish_latency_ms"

	METRIC_TARGET_MONGO = "mongo"
	METRIC_NAME_QUERY_LATENCY = "query_latency_ms"
	METRIC_NAME_QUERY_RATE = "query_rate"

	METRIC_TARGET_FIREBASE = "firebase"
	METRIC_NAME_MULTICAST_LATENCY = "multicast_latency_ms"

	METADATA_AGGREGATE_FLAG = "should_aggregate"
	METADATA_AGGREGATION_ACCUMULATE = "aggregation_accumulate"
	METADATA_AGGREGATION_MAXIMUM = "aggregation_maximum"

	METADATA_RATE_FLAG = "should_compute_rate"
)