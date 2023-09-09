package publish

type Publish interface {
	// Send sends data to Broker.
	// Note that data is recommended to design into a struct include a string key and a bytes type data.
	//
	// Suggestion: Use merge.Merge to merge same events in a tiny interval.
	Send(data []byte) error
	Run() error
	Close() error
}
