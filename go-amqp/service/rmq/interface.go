package rmq

type Receiver interface {
	QueueName() string
	RouterKey() string
	OnError(error)
	OnReceive([]byte) bool
}
