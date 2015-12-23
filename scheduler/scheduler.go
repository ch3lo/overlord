package scheduler

// Scheduler es una interfaz que debe implementar cualquier scheduler de Servicios
// Para un ejemplo ir a swarm.Scheduler
type Scheduler interface {
	ID() string
	IsAlive(id string) (bool, error)
	Instances() ([]ServiceInformation, error)
}
