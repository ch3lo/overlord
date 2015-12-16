package service

import "fmt"

// AlreadyExist sucede cuando se esta agregando un servicio que ya existe
type AlreadyExist struct {
	Name string
}

func (err AlreadyExist) Error() string {
	return fmt.Sprintf("El servicio ya existe: %s", err.Name)
}

// ManagerAlreadyExist sucede cuando se esta agregando un manager de servicio que ya existe
type ManagerAlreadyExist struct {
	Service string
	Version string
}

func (err ManagerAlreadyExist) Error() string {
	return fmt.Sprintf("El manager %s del servicio %s ya existe", err.Version, err.Service)
}

// ImageNameRegexpError se lanza cuando no se puede compilar el nombre de la imagen como expresion regular
type ImageNameRegexpError struct {
	Regexp  string
	Message string
}

func (err ImageNameRegexpError) Error() string {
	return fmt.Sprintf("No se pudo compilar el nombre de la imagen %s como expresion regular: %s", err.Regexp, err.Message)
}
