package scheduler

// ServiceInformationStatus define el estado de un servicio
type ServiceInformationStatus int

const (
	// ServiceUp Estado de un servicio que está OK
	ServiceUp ServiceInformationStatus = 1 + iota
	// ServiceDown Estado de un servicio que esta caido
	ServiceDown
)

var statuses = [...]string{
	"ServiceUp",
	"ServiceDown",
}

func (s ServiceInformationStatus) String() string {
	return statuses[s-1]
}

// ServiceInformation define una estructura con la informacion basica de un servicio
// Esta estructura sirve para la comunicacion con los consumidores de schedulers
type ServiceInformation struct {
	ID            string
	ImageName     string
	ImageTag      string
	Host          string
	ContainerName string
	Status        ServiceInformationStatus
}

// Healthy es una funcion que retorna si un servicio esta saludable o no
func (si ServiceInformation) Healthy() bool {
	return si.Status == ServiceUp
}
