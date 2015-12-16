package factory

import (
	"fmt"

	"github.com/ch3lo/overlord/notification"
	"github.com/ch3lo/overlord/util"
)

// notificationFactories almacena una mapeo entre un typo de notificador y su constructor
var notificationFactories = make(map[string]NotificationFactory)

// NotificationFactory es una interfaz para crear un Notificador
// Cada Notificador debe implementar estar interfaz y además llamar el metodo Register
type NotificationFactory interface {
	Create(id string, parameters map[string]interface{}) (notification.Notification, error)
}

// Register permite a una implementación de Notification estar disponible mediante un id que representa el tipo de notificador
func Register(name string, factory NotificationFactory) {
	if factory == nil {
		util.Log.Fatal("Se debe pasar como argumento un NotificationFactory")
	}
	_, registered := notificationFactories[name]
	if registered {
		util.Log.Fatalf("NotificationFactory %s ya está registrado", name)
	}

	notificationFactories[name] = factory
}

// Create crea un notificador a partir del tipo de notificacion.
// Si el notificador no estaba registrado se retornará un InvalidNotification
func Create(name string, id string, parameters map[string]interface{}) (notification.Notification, error) {
	notificationFactory, ok := notificationFactories[name]
	if !ok {
		return nil, InvalidNotification{name}
	}
	return notificationFactory.Create(id, parameters)
}

// InvalidNotification sucede cuando se instenta construir un Notificador no registrado
type InvalidNotification struct {
	Name string
}

func (err InvalidNotification) Error() string {
	return fmt.Sprintf("Notificador no esta registrado: %s", err.Name)
}
