package email

import (
	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
)

const notificationId = "email"

func init() {
	factory.Register(notificationId, &emailCreator{})
}

// emailCreator implementa la interfaz factory.NotificationFactory
type emailCreator struct{}

func (factory *emailCreator) Create(parameters map[string]interface{}) (notification.Notification, error) {
	return NewFromParameters(parameters)
}

// EmailParameters encapsula los parametros de configuracion de Email
type EmailParameters struct {
	from     string
	subject  string
	smtp     string
	user     string
	password string
}

// NewFromParameters construye un EmailNotification a partir de un mapeo de parámetros
func NewFromParameters(parameters map[string]interface{}) (*EmailNotification, error) {
	params := EmailParameters{}
	return New(params)
}

// New construye un nuevo EmailNotification
func New(params EmailParameters) (*EmailNotification, error) {

	email := new(EmailNotification)

	return email, nil
}

// EmailNotification es una implementacion de notification.Notification
// Permite la comunicacion via email
type EmailNotification struct {
}

func (n *EmailNotification) Id() string {
	return notificationId
}

func (n *EmailNotification) Notify() {
}
