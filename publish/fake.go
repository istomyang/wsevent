package publish

import (
	"github.com/istomyang/wsevent/log"
)

type fakePublisher struct{}

func NewFakePublisher() Publish {
	return &fakePublisher{}
}

func (f *fakePublisher) Send(data []byte) error {
	log.Debug("fakePublisher-send-data: %s", string(data))
	return nil
}

func (f *fakePublisher) Run() error {
	log.Debug("fakePublisher: run")
	return nil
}

func (f *fakePublisher) Close() error {
	log.Debug("fakePublisher: close")
	return nil
}

var _ Publish = &fakePublisher{}
