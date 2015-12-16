package monitor

import "regexp"

type ServiceChangeCriteria interface {
	MeetCriteria(status map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData
}

type ImageNameAndImageTagRegexpCriteria struct {
	FullImageNameRegexp *regexp.Regexp
}

func (c *ImageNameAndImageTagRegexpCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	for k, v := range elements {
		fullName := v.origin.ImageName + ":" + v.origin.ImageTag
		if c.FullImageNameRegexp.MatchString(fullName) {
			filtered[k] = elements[k]
		}
	}
	return filtered
}

type StatusCriteria struct {
	Status ServiceDataStatus
}

func (c *StatusCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	for k, v := range elements {
		if v.InStatus(c.Status) {
			filtered[k] = elements[k]
		}
	}
	return filtered
}

type HealthyCriteria struct {
	Status bool
}

func (c *HealthyCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	for k, v := range elements {
		if v.Origin().Healthy() == c.Status {
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
