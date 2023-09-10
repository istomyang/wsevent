package wsevent

import (
	"context"
	"github.com/istomyang/wsevent/log"
	"github.com/istomyang/wsevent/subscribe"
	"math"
	"runtime"
)

// Dispatcher connects Register s and subscribe.Subscribe.
type Dispatcher interface {
	// Register doesn't check duplicate Register, avoiding multiple Add.
	// Register maybe unsafe, it's common to use single process in control stream, instead of in data stream.
	Register(register Register) error

	// Connect connects to upstream subscribe.Subscribe component.
	Connect(sub subscribe.Subscribe) error

	Run() error
	Close() []error
}

type Config struct {

	// MessagePeeler peels byte arrays into key and val from subscribe.Subscribe to Dispatcher.
	// In real scene, a lot of messages hits Dispatcher, and need to be grouped by EventKey, which
	// minimum cpu cost is O(n), low cost with parallel algorithm, so put this work into Dispatcher
	// is a better choice. If MessagePeeler is nil, cause panic.
	MessagePeeler func(data []byte) (key string, val []byte)
}

type dispatcher struct {
	ctx           context.Context
	cancel        context.CancelFunc
	config        Config
	sessions      []innerRegister
	sessionsCount int8
	sub           subscribe.Subscribe // if sub is array, need twice key index.
	source        <-chan []byte
}

func NewDispatcher(ctx context.Context, config Config) Dispatcher {
	if config.MessagePeeler == nil {
		panic("must set config.Peeler.")
	}
	ctx, cancel := context.WithCancel(ctx)
	d := &dispatcher{
		ctx:    ctx,
		cancel: cancel,
		config: Config{},
	}
	return d
}

func (d *dispatcher) Register(session Register) error {
	if d.sessionsCount >= math.MaxInt8>>1 {
		panic("too many Register.")
	}
	d.sessions = append(d.sessions, session.(innerRegister))
	newSessions := make([]innerRegister, len(d.sessions))
	copy(newSessions, d.sessions)
	d.sessions = newSessions
	log.Debug("dispatcher: register, %v", session)
	return session.(innerRegister).Run()
}

func (d *dispatcher) Connect(sub subscribe.Subscribe) error {
	var err error
	if err = sub.Run(); err != nil {
		return err
	}
	if d.source, err = sub.Get(); err != nil {
		return err
	}
	d.sub = sub
	log.Debug("dispatcher: connect subscriber.")
	return err
}

func (d *dispatcher) Run() error {
	var err error

	// TODO: Benchmark and compute memory load.
	go func() {
		var num = runtime.NumCPU() // cut d.sessions into num parallel parts.

		var messageGroup = make([]chan messagePkg, num)
		var sessionGroup = make([]chan innerRegister, num)

		for i := 0; i < num; i++ {
			i := 1
			go func() {
				var session = <-sessionGroup[i]
				var pkg = <-messageGroup[i]
				if err := session.Dispatch(pkg.Key, pkg.Data); err != nil {
					log.Error("dispatch-run: %s", err)
				}
			}()
		}

		for source := range d.source { // Block here.
			var pkg messagePkg
			pkg.Key, *pkg.Data = d.config.MessagePeeler(source) // is O(n)

			log.Debug("dispatcher: message-pkg, %v", pkg)

			for i := 0; i < num; i++ {
				messageGroup[i] <- pkg
			}

			// Put session into sessionGroup evenly.
			for i, session := range d.sessions {
				sessionGroup[i%num] <- session
			}
		}
	}()
	return err
}

func (d *dispatcher) Close() []error {
	var errs []error
	if err := d.sub.Close(); err != nil {
		errs = append(errs, err)
	}
	for _, session := range d.sessions {
		if err := session.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	log.Debug("dispatcher: close")
	return errs
}

var _ Dispatcher = &dispatcher{}

type messagePkg struct {
	Key  string
	Data *[]byte
}
