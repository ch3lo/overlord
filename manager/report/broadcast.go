package report

import (
	"time"

	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/util"
	"gopkg.in/matryer/try.v1"
)

type Broadcast interface {
	Broadcast(data []byte)
	Register(n notification.Notification) error
}

type Broadcaster struct {
	attemptsOnError  int
	waitOnError      time.Duration
	waitAfterAttemts time.Duration
	workers          map[string]notification.Notification
}

func NewBroadcaster(attemptsOnError int, waitOnError time.Duration, waitAfterAttemts time.Duration) *Broadcaster {
	if attemptsOnError == 0 {
		attemptsOnError = 5
	}

	if waitOnError == 0 {
		waitOnError = 30 * time.Second
	}

	if waitAfterAttemts == 0 {
		waitAfterAttemts = 60 * time.Second
	}

	b := &Broadcaster{
		attemptsOnError:  attemptsOnError,
		waitOnError:      waitOnError,
		waitAfterAttemts: waitAfterAttemts,
		workers:          make(map[string]notification.Notification),
	}

	return b
}

func (b *Broadcaster) Register(n notification.Notification) error {
	if _, ok := b.workers[n.ID()]; ok {
		return &BroadcastWorkerAlreadyExist{Name: n.ID()}
	}
	b.workers[n.ID()] = BroadcastWorker{
		attemptsOnError:  b.attemptsOnError,
		waitOnError:      b.waitOnError,
		waitAfterAttemts: b.waitAfterAttemts,
		notification:     n,
	}
	return nil
}

func (b *Broadcaster) Broadcast(data []byte) {
	for _, v := range b.workers {
		v.Notify(data)
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
	attemptsOnError  int
	waitOnError      time.Duration
	waitAfterAttemts time.Duration
	notification     notification.Notification
	status           boradcastStatus
	quitChan         chan bool
}

func (w BroadcastWorker) ID() string {
	return w.notification.ID()
}

func (w BroadcastWorker) Notify(data []byte) error {
	w.status.total++

	go func() {
		for {
			select {
			case <-w.quitChan:
				util.Log.Infoln("Notificacion detenida")
				return

			default:
				err := try.Do(func(attempt int) (bool, error) {
					err := w.notification.Notify(data)
					if err == nil {
						return false, nil
					}
					w.status.errors++
					retry := attempt < w.attemptsOnError
					if retry {
						time.Sleep(w.waitOnError)
					}
					return retry, err
				})
				if err == nil {
					w.status.success++
					return
				}
				w.status.fail++
				util.Log.Warnf("No se pudo notificar, se esperara un tiempo: %s", err.Error())
				time.Sleep(w.waitAfterAttemts)
			}
		}
	}()
	return nil
}
