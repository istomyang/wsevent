package subscribe

import "wsevent/log"

type fakeSubscriber struct {
	config FakeConfig
}

type FakeConfig struct {
	PublishSend <-chan []byte
}

func NewFakeSubscriber(config FakeConfig) Subscribe {
	return &fakeSubscriber{config: config}
}

func (f *fakeSubscriber) Get() (<-chan []byte, error) {
	return f.config.PublishSend, nil
}

func (f *fakeSubscriber) Run() error {
	log.Debug("fakeSubscriber: run")
	return nil
}

func (f *fakeSubscriber) Close() error {
	log.Debug("fakeSubscriber: close")
	return nil
}

var _ Subscribe = &fakeSubscriber{}
