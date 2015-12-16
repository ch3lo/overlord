package scheduler

type ServiceInformationStatus int

const (
	SERVICE_UP ServiceInformationStatus = 1 + iota
	SERVICE_DOWN
)

var statuses = [...]string{
	"SERVICE_UP",
	"SERVICE_DOWN",
}

func (s ServiceInformationStatus) String() string {
	return statuses[s-1]
}

type ServiceInformation struct {
	Id         string
	ImageName  string
	ImageTag   string
	Status     ServiceInformationStatus
	FullStatus string
}

func (si ServiceInformation) Healthy() bool {
	return si.Status == SERVICE_UP
}
