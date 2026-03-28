package zkillboard

import (
	"log/slog"
	"sync"
)

type killmailPropagator struct {
	channels              []chan *Killmail
	channelsMutex         sync.Mutex
	externalSenderChannel chan *Killmail
	started               bool
}

var KillmailPropagator killmailPropagator

func (propagator *killmailPropagator) New(inputChannel chan *Killmail) killmailPropagator {
	return killmailPropagator{
		externalSenderChannel: inputChannel,
	}
}

func (propagator *killmailPropagator) Start() {
	propagator.channelsMutex.Lock()
	defer propagator.channelsMutex.Unlock()

	if propagator.started {
		slog.Warn("Tried to start an already started killmail propagator")
		return
	}

	go func() {
		for {
			killmail := <-propagator.externalSenderChannel

			propagator.channelsMutex.Lock()
			for _, channel := range propagator.channels {
				select {
				case channel <- killmail:
				default:
					slog.Warn("Receiver channel was full, dropping message")
				}
			}
			propagator.channelsMutex.Unlock()
		}
	}()
}

func (propagator *killmailPropagator) AddListenerChannel(channel chan *Killmail) {
	propagator.channelsMutex.Lock()
	propagator.channels = append(propagator.channels, channel)
	propagator.channelsMutex.Unlock()
}
