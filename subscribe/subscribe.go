package subscribe

type Subscribe interface {
	// Get use key to recognize messages what I need.
	// Note that data is recommended to design into a struct include a string key and a bytes type data.
	Get() (<-chan []byte, error)
	Run() error
	Close() error
}
