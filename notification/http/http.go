package http

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ch3lo/overlord/logger"
	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/notification/factory"
)

const notificationID = "http"

func init() {
	factory.Register(notificationID, &httpCreator{})
}

// httpCreator implementa la interfaz factory.NotificationFactory
type httpCreator struct{}

func (factory *httpCreator) Create(id string, params map[string]interface{}) (notification.Notification, error) {
	return NewFromParameters(id, params)
}

// parameters encapsula los parametros de configuracion de Email
type parameters struct {
	id      string
	url     string
	headers string
	method  string
}

// NewFromParameters construye un Notification a partir de un mapeo de parámetros
func NewFromParameters(id string, params map[string]interface{}) (*Notification, error) {

	url, ok := params["url"]
	if !ok || fmt.Sprint(url) == "" {
		return nil, errors.New("Parametro url no existe")
	}

	method, ok := params["method"]
	if !ok || fmt.Sprint(method) == "" {
		return nil, errors.New("Parametro method no existe")
	}

	p := parameters{
		id:     id,
		url:    fmt.Sprint(url),
		method: fmt.Sprint(method),
	}
	return New(p)
}

// New construye un nuevo Notification
func New(params parameters) (*Notification, error) {

	http := &Notification{
		id:     params.id,
		url:    params.url,
		method: params.method,
	}

	return http, nil
}

// Notification es una implementacion de notification.Notification
// Permite la comunicacion via email
type Notification struct {
	id     string
	url    string
	method string
}

// ID retorna el identificador de este notificador
func (n *Notification) ID() string {
	return n.id
}

// Notify notifica via http al endpoint configurado
func (n *Notification) Notify(data []byte) error {
	logger.Instance().Infoln("Notificando via http")
	logger.Instance().Debugf("Data: %s", string(data))

	//var query = []byte(`your query`)
	req, err := http.NewRequest("POST", n.url, bytes.NewBuffer(data))
	//req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logger.Instance().Debugf("Response Status: %s - Header: %s", resp.Status, resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	logger.Instance().Debugf("Response Body: %s", string(body))
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New(fmt.Sprintf("Respuesta con estado invalido %s", resp.Status))
}
