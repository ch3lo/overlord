package monitor

import "regexp"

// ServiceChangeCriteria esta interfaz sirve para crear filtros que se aplican a ServiceUpdaterData
type ServiceChangeCriteria interface {
	MeetCriteria(status map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData
}

// ImageNameAndImageTagRegexpCriteria filtra aquellos servicios cuyo nombre y tagde imagen no
// cumplen con la expresion regular FullImageNameRegexp
type ImageNameAndImageTagRegexpCriteria struct {
	FullImageNameRegexp *regexp.Regexp
}

// MeetCriteria aplica el filtro ImageNameAndImageTagRegexpCriteria y retorna un map[string]*ServiceUpdaterData
// con aquellos servicios que cumplen con el criterio
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

// StatusCriteria es un filtro que aplica criterios sobre el estado de un servicio
type StatusCriteria struct {
	Status ServiceDataStatus
}

// MeetCriteria aplica el filtro StatusCriteria y retorna un map[string]*ServiceUpdaterData
// con aquellos servicios que cumplen con el criterio
func (c *StatusCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	for k, v := range elements {
		if v.InStatus(c.Status) {
			filtered[k] = elements[k]
		}
	}
	return filtered
}

// HealthyCriteria es un filtro que aplica criterios sobre la salud de un servicio
type HealthyCriteria struct {
	Status bool
}

// MeetCriteria aplica el filtro HealthyCriteria y retorna un map[string]*ServiceUpdaterData
// con aquellos servicios que cumplen con el criterio
func (c *HealthyCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := make(map[string]*ServiceUpdaterData)
	/*	for k, v := range elements {
		if v.Origin().Healthy() == c.Status {
			filtered[k] = elements[k]
		}
	}*/
	return filtered
}

// AndCriteria es un criterio que se puede aplicar para realizar un && sobre otros dos criterios
type AndCriteria struct {
	criteria      ServiceChangeCriteria
	otherCriteria ServiceChangeCriteria
}

// MeetCriteria aplica el filtro que tiene como objetivo realizar un && sobre dos criterios
func (c *AndCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := c.criteria.MeetCriteria(elements)
	return c.otherCriteria.MeetCriteria(filtered)
}

// OrCriteria es un criterio que se puede aplicar para realizar un || sobre otros dos criterios
type OrCriteria struct {
	criteria      ServiceChangeCriteria
	otherCriteria ServiceChangeCriteria
}

// MeetCriteria aplica el filtro que tiene como objetivo realizar un || sobre dos criterios
func (c *OrCriteria) MeetCriteria(elements map[string]*ServiceUpdaterData) map[string]*ServiceUpdaterData {
	filtered := c.criteria.MeetCriteria(elements)
	others := c.otherCriteria.MeetCriteria(elements)

	for k := range others {
		if filtered[k] == nil {
			filtered[k] = others[k]
		}
	}

	return filtered
}
