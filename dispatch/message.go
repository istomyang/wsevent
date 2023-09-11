package dispatch

type Message struct {
	Key  EventKey
	Data []byte
}
