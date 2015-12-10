package monitor

type ServiceChangeCriteria interface {
	MeetCriteria(status map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData
}

type ImageNameCriteria struct {
	Name string
}

func (c *ImageNameCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	for k, v := range elements {
		if c.Name == v.origin.Image {
			filtered[k] = elements[k]
		}
	}
	return filtered
}

type StatusCriteria struct {
	Status string
}

func (c *StatusCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	for k, v := range elements {
		if c.Status == v.origin.Status {
			filtered[k] = elements[k]
		}
	}
	return filtered
}

type AndCriteria struct {
	criteria      ServiceChangeCriteria
	otherCriteria ServiceChangeCriteria
}

func (c *AndCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := c.criteria.MeetCriteria(elements)
	return c.otherCriteria.MeetCriteria(filtered)
}

type OrCriteria struct {
	criteria      ServiceChangeCriteria
	otherCriteria ServiceChangeCriteria
}

func (c *OrCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := c.criteria.MeetCriteria(elements)
	others := c.otherCriteria.MeetCriteria(elements)

	for k, _ := range others {
		if filtered[k] == nil {
			filtered[k] = others[k]
		}
	}

	return filtered
}
