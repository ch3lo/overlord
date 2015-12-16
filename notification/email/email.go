package email

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
	"github.com/ch3lo/overlord/util"
)

const notificationId = "email"

func init() {
	factory.Register(notificationId, &emailCreator{})
}

// emailCreator implementa la interfaz factory.NotificationFactory
type emailCreator struct{}

func (factory *emailCreator) Create(id string, parameters map[string]interface{}) (notification.Notification, error) {
	return NewFromParameters(id, parameters)
}

// EmailParameters encapsula los parametros de configuracion de Email
type EmailParameters struct {
	id       string
	from     string
	subject  string
	smtp     string
	user     string
	password string
}

// NewFromParameters construye un EmailNotification a partir de un mapeo de parámetros
func NewFromParameters(id string, parameters map[string]interface{}) (*EmailNotification, error) {

	smtp, ok := parameters["smtp"]
	if !ok || fmt.Sprint(smtp) == "" {
		return nil, errors.New("Parametro smtp no existe")
	}

	subject := ""
	if subjectTmp, ok := parameters["subject"]; ok {
		subject = fmt.Sprint(subjectTmp)
	}

	params := EmailParameters{
		id:      id,
		smtp:    fmt.Sprint(smtp),
		subject: subject,
	}
	return New(params)
}

// New construye un nuevo EmailNotification
func New(params EmailParameters) (*EmailNotification, error) {

	email := &EmailNotification{
		id: params.id,
	}

	return email, nil
}

// EmailNotification es una implementacion de notification.Notification
// Permite la comunicacion via email
type EmailNotification struct {
	id string
}

func (n *EmailNotification) Id() string {
	return n.id
}

func (n *EmailNotification) Notify() error {
	util.Log.Infoln("Notificando via email")
	if rand.Int()%2 == 0 {
		util.Log.Errorln("Error en notificacion")
		return errors.New("Random gatillo error de email")
	}
	return errors.New("ERrro forzado")
}
