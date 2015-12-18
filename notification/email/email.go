package email

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"

	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
	"github.com/ch3lo/overlord/util"
)

const notificationID = "email"

func init() {
	factory.Register(notificationID, &emailCreator{})
}

// emailCreator implementa la interfaz factory.NotificationFactory
type emailCreator struct{}

func (factory *emailCreator) Create(id string, params map[string]interface{}) (notification.Notification, error) {
	return NewFromParameters(id, params)
}

// EmailParameters encapsula los parametros de configuracion de Email
type parameters struct {
	id       string
	from     string
	to       string
	subject  string
	smtp     string
	user     string
	password string
}

// NewFromParameters construye un EmailNotification a partir de un mapeo de parámetros
func NewFromParameters(id string, params map[string]interface{}) (*Notification, error) {

	smtp, ok := params["smtp"]
	if !ok || fmt.Sprint(smtp) == "" {
		return nil, errors.New("Parametro smtp no existe")
	}

	from, ok := params["from"]
	if !ok || fmt.Sprint(from) == "" {
		return nil, errors.New("Parametro de origen (from) no existe")
	}

	to, ok := params["to"]
	if !ok || fmt.Sprint(to) == "" {
		return nil, errors.New("Parametro de destinatario (to) no existe")
	}

	// TODO implementacion con autenticacion y subject
	/*
		subject := ""
		if subjectTmp, ok := parameters["subject"]; ok {
			subject = fmt.Sprint(subjectTmp)
		}*/

	p := parameters{
		id:   id,
		smtp: fmt.Sprint(smtp),
		from: fmt.Sprint(from),
		to:   fmt.Sprint(to),
	}
	return New(p)
}

// New construye un nuevo EmailNotification
func New(params parameters) (*Notification, error) {

	email := &Notification{
		id:      params.id,
		address: params.smtp,
		from:    params.from,
		to:      params.to,
	}

	return email, nil
}

// Notification es una implementacion de notification.Notification
// Permite la comunicacion via email
type Notification struct {
	id      string
	address string
	from    string
	to      string
}

// ID retorna el identificador de este notificador
func (n *Notification) ID() string {
	return n.id
}

// Notify notifica via email al destinatario
func (n *Notification) Notify(data []byte) error {
	util.Log.Infoln("Notificando via email")
	util.Log.Debugf("Data: %s", string(data))

	// Connect to the remote SMTP server.
	c, err := smtp.Dial(n.address)
	if err != nil {
		return err
	}
	defer c.Close()

	// Set the sender and recipient.
	c.Mail(n.from)
	c.Rcpt(n.to)

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	buf := bytes.NewBuffer(data)
	//buf := bytes.NewBufferString("This is the email body.")
	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}

	return nil
}
