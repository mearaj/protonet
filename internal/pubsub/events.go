package pubsub

import (
	"errors"
	model2 "github.com/mearaj/protonet/internal/model"
	"github.com/mearaj/protonet/utils"
	"sync"
)

type Topic int

const (
	SendNewMessageEventTopic Topic = iota
	CurrentAccountChangedEventTopic
	AccountsChangedEventTopic
	ContactsChangedEventTopic
	MessageStateChangedEventTopic
	MessagesStateChangedEventTopic
	UserPasswordChangedEventTopic
	GetContactsEventTopic
	GetMessagesTopic
	SendNewMessageTopic
	SaveContactTopic
	NewMessageReceivedTopic
	DatabaseOpened
)

var AllTopicsArr = [...]Topic{
	SendNewMessageEventTopic,
	CurrentAccountChangedEventTopic,
	AccountsChangedEventTopic,
	ContactsChangedEventTopic,
	MessageStateChangedEventTopic,
	MessagesStateChangedEventTopic,
	UserPasswordChangedEventTopic,
	GetContactsEventTopic,
	GetMessagesTopic,
	SendNewMessageTopic,
	SaveContactTopic,
	NewMessageReceivedTopic,
	DatabaseOpened,
}

type DatabaseOpenedEventData struct{}
type AccountsChangedEventData struct{}
type CurrentAccountChangedEventData struct {
	PrevAccountPublicKey    string
	CurrentAccountPublicKey string
}
type ContactsChangeEventData struct{ AccountPublicKey string }
type MessagesStateChangedEventData struct {
	AccountPublicKey string
	ContactPublicKey string
}
type MessageStateChangedEventData struct {
	model2.Message
}
type NewMessageReceivedEventData struct {
	model2.Message
}

type SendNewMessageEventData struct {
	model2.Message
}
type SaveContactEventData struct {
	model2.Contact
}

type Event struct {
	Data   interface{}
	Topic  Topic
	Cached bool
	Err    error
}

type EventCallback func(event Event)

type Subscriber struct {
	events      chan Event
	topics      utils.Map[Topic, struct{}]
	closed      bool
	closedMutex sync.RWMutex
	callback    EventCallback
}
type Subscription interface {
	// Events the channel where Event can be received,
	Events() <-chan Event
	// Subscribe should subscribe to all events if empty, can return error esp when subscription is closed
	Subscribe(...Topic) error
	// IsSubscribedTo returns bool indicating Subscriber's subscription status to a topic
	//  can return error esp when subscription is closed
	IsSubscribedTo(topic Topic) (bool, error)
	// UnSubscribe should unsubscribe to all events if empty, the Subscriber should still be open/usable
	//  can return error esp when subscription is closed
	UnSubscribe(...Topic) error
	// Close should make this interface unusable and release resources
	Close()
	// IsClosed indicates this interface is still usable
	IsClosed() bool
	// Topics returns topics to which Subscriber is subscribed to,
	// can return error esp when subscription is closed
	Topics() ([]Topic, error)
	// SubscribeWithCallback (optionally listening with callback),
	// can return error esp when subscription is closed
	SubscribeWithCallback(callback EventCallback)
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		events: make(chan Event, 100),
		closed: false,
		topics: utils.NewMap[Topic, struct{}](),
	}
}

func (s *Subscriber) Events() <-chan Event {
	s.closedMutex.RLock()
	defer s.closedMutex.RUnlock()
	return s.events
}

func (s *Subscriber) isError() error {
	if s.IsClosed() {
		return errors.New("subscription already closed")
	}
	return nil
}

// Subscribe subscribes to all events if empty
func (s *Subscriber) Subscribe(topics ...Topic) (err error) {
	if err = s.isError(); err != nil {
		return err
	}
	if len(topics) == 0 {
		for _, eachTopic := range AllTopicsArr {
			s.topics.Set(eachTopic, struct{}{})
		}
		return nil
	}

	for _, eachTopic := range topics {
		s.topics.Set(eachTopic, struct{}{})
	}
	return nil
}

func (s *Subscriber) IsSubscribedTo(topic Topic) (ok bool, err error) {
	if err = s.isError(); err != nil {
		return ok, err
	}
	_, ok = s.topics.Get(topic)
	return ok, err
}

// UnSubscribe unsubscribes to all events if empty, Subscriber is still referenced
func (s *Subscriber) UnSubscribe(topics ...Topic) (err error) {
	if err = s.isError(); err != nil {
		return err
	}
	if len(topics) == 0 {
		s.topics.Clear()
		return err
	}

	for _, eachTopic := range topics {
		s.topics.Delete(eachTopic)
	}
	return err
}

// Close closes the Event chan and UnSubscribes to all events and clears EventCallback
func (s *Subscriber) Close() {
	_ = s.UnSubscribe()
	s.closedMutex.Lock()
	s.closed = true
	s.closedMutex.Unlock()
	s.callback = nil
	go close(s.events)
}
func (s *Subscriber) IsClosed() bool {
	s.closedMutex.RLock()
	defer s.closedMutex.RUnlock()
	return s.closed
}

func (s *Subscriber) Topics() (topics []Topic, err error) {
	if err = s.isError(); err != nil {
		return nil, err
	}
	return s.topics.Keys(), err
}
func (s *Subscriber) fire(event Event) {
	if ok, err := s.IsSubscribedTo(event.Topic); ok && err == nil && !s.IsClosed() {
		select {
		case s.events <- event:
		default:
			// channel buffer is full, empty first element and append to last element
			select {
			case <-s.Events():
				s.events <- event
			default:
			}
		}
		if s.callback != nil {
			s.callback(event)
		}
	}
}
func (s *Subscriber) SubscribeWithCallback(callback EventCallback) {
	s.callback = callback
}

type subscribers = utils.Map[*Subscriber, struct{}]

type EventBroker struct {
	cachedEvents utils.Map[Topic, Event]
	subscribers  subscribers
}

func NewEventBroker() *EventBroker {
	return &EventBroker{
		cachedEvents: utils.NewMap[Topic, Event](),
		subscribers:  utils.NewMap[*Subscriber, struct{}](),
	}
}

func (eb *EventBroker) AddSubscriber(sub *Subscriber) {
	eb.subscribers.Set(sub, struct{}{})
	for _, e := range eb.cachedEvents.Values() {
		if ok, err := sub.IsSubscribedTo(e.Topic); ok && err == nil {
			go sub.fire(e)
		}
	}
}

func (eb *EventBroker) Fire(event Event) {
	for _, sub := range eb.subscribers.Keys() {
		eb.cachedEvents.Set(event.Topic, event)
		if sub.IsClosed() {
			eb.subscribers.Delete(sub)
			continue
		}
		go sub.fire(event)
	}
}

func AddSubscriber(eventBroker *EventBroker, topic ...Topic) Subscription {
	subscription := NewSubscriber()
	_ = subscription.Subscribe(topic...)
	eventBroker.AddSubscriber(subscription)
	return subscription
}
