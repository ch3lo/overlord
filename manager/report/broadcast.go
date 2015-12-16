package report

import "github.com/ch3lo/overlord/notification"

type Broadcast interface {
	Broadcast()
	Register(n notification.Notification) error
}

type Broadcaster struct {
	notifications map[string]notification.Notification
}

func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		notifications: make(map[string]notification.Notification),
	}

	return b
}

func (b *Broadcaster) Register(n notification.Notification) error {
	if _, ok := b.notifications[n.Id()]; ok {
		return &BroadcasterAlreadyExist{Name: n.Id()}
	}
	b.notifications[n.Id()] = n
	return nil
}

func (b *Broadcaster) Broadcast() {
	for _, v := range b.notifications {
		v.Notify()
	}
}
