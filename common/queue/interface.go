package queue

type (
	// Queue is a general queue interface
	Queue interface {
		Publish(msgs interface{}) error
		Close() error
	}
)
