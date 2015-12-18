package notification

// Notification es una interfaz que deben implementar los notificadores
// Para un ejemplo ir a notification.Email
type Notification interface {
	ID() string
	Notify(data []byte) error
}
