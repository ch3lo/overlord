package monitor

// ServiceUpdaterSubscriber es una interfaz que deben implementar aquellos subscriptores
// que desean recibir notificacions de los servicios manejados por ServiceUpdater
type ServiceUpdaterSubscriber interface {
	ID() string
	Update(map[string]*ServiceUpdaterData)
}
