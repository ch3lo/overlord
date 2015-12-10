package monitor

type ServiceUpdaterSubscriber interface {
	Id() string
	Update(map[string]*ServiceUpdaterData)
}
