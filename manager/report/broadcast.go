package report

import (
	"time"

	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/util"
	"gopkg.in/matryer/try.v1"
)

type Broadcast interface {
	Broadcast()
	Register(n notification.Notification) error
}

type Broadcaster struct {
	workers map[string]notification.Notification
}

func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		workers: make(map[string]notification.Notification),
	}

	return b
}

func (b *Broadcaster) Register(n notification.Notification) error {
	if _, ok := b.workers[n.Id()]; ok {
		return &BroadcastWorkerAlreadyExist{Name: n.Id()}
	}
	b.workers[n.Id()] = BroadcastWorker{notification: n}
	return nil
}

func (b *Broadcaster) Broadcast() {
	for _, v := range b.workers {
		v.Notify()
	}
}

func (b *Broadcaster) Send() {

}

type boradcastStatus struct {
	total   int
	success int
	errors  int
	fail    int
}

type BroadcastWorker struct {
	notification notification.Notification
	status       boradcastStatus
	quitChan     chan bool
}

func (w BroadcastWorker) Id() string {
	return w.notification.Id()
}

func (w BroadcastWorker) Notify() error {
	w.status.total++

	go func() {
		for {
			select {
			case <-w.quitChan:
				util.Log.Infoln("Notificacion detenida")
				return

			default:
				err := try.Do(func(attempt int) (bool, error) {
					err := w.notification.Notify()
					if err == nil {
						return false, nil
					}
					w.status.errors++
					return attempt < 5, err
				})
				if err == nil {
					w.status.success++
					return
				}
				w.status.fail++
				util.Log.Warnf("No se pudo notificar, se esperara un tiempo: %s", err.Error())
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return nil
}
