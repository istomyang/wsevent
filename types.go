package wsevent

type EventKey = string

type EventHandler = func(req []byte) ([]byte, error)

type void struct{}

var noop void
