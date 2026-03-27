package feedboard

const feedboardMessageChannelBuffer = 16

type feedboardSubscriber struct {
	MessagesChannel chan []byte
	CloseSlowFunc   func()
}

type feedboardSubscriberBuilder struct{}

var FeedboardSubscriberBuilder feedboardSubscriberBuilder

func (builder *feedboardSubscriberBuilder) NewSubscriber(closeSlowFunc func()) feedboardSubscriber {
	return feedboardSubscriber{
		MessagesChannel: make(chan []byte, feedboardMessageChannelBuffer),
		CloseSlowFunc:   closeSlowFunc,
	}
}
