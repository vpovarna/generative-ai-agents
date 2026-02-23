package stream

type StreamConfig struct {
	RedisAddr    string
	Stream       string
	Group        string
	ConsumerName string
}

func NewStreamConfig(redisAddr string, stream string, group string, consumerName string) *StreamConfig {
	return &StreamConfig{
		RedisAddr:    redisAddr,
		Stream:       stream,
		Group:        group,
		ConsumerName: consumerName,
	}
}
