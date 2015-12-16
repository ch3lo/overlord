package scheduler

// Scheduler es una interfaz que debe implementar cualquier scheduler de Servicios
// Para un ejemplo ir a swarm.SwarmScheduler
type Scheduler interface {
	ID() string
	IsAlive(id string) (bool, error)
	Instances(filter FilterInstances) ([]ServiceInformation, error)
}

// FilterInstances es una estrutura para encapsular los requerimientos
// que se utilizaran para filtrar instancias de servicios
type FilterInstances struct {
	imageName string
	imageTag  string
}
