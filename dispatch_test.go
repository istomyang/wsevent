package wsevent

import (
	"encoding/json"
)

// ClientRequestBody defines ws.Client's request format.
type ClientRequestBody struct {
	Key  EventKey
	Body []byte
}

func (req ClientRequestBody) Success() ClientResponseBody {
	return ClientResponseBody{
		Key:  "reply_success",
		Body: []byte("some messages"),
	}
}

func (req ClientRequestBody) Fail() ClientResponseBody {
	var body = struct {
		Code    string
		Message string
	}{Code: "001201", Message: "some error occurs."}

	data, _ := json.Marshal(body)

	return ClientResponseBody{
		Key:  "reply_success",
		Body: data,
	}
}

// ClientResponseBody defines ws.Server's response format.
type ClientResponseBody struct {
	Key  EventKey
	Body []byte
}

// MessageBody defines publish.Publish and subscribe.Subscribe 's message format.
type MessageBody struct {
	Key  EventKey
	Body []byte
}
