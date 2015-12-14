package cluster

import "fmt"

// ClusterDoesntExits error generado cuando un cluster no existe
type ClusterDoesntExits struct {
	Name string
}

func (err ClusterDoesntExits) Error() string {
	return fmt.Sprintf("El cluster no existe: %s", err.Name)
}

// ClusterDisabled error generado cuando un cluster no esta habilitado
type ClusterDisabled struct {
	Name string
}

func (err ClusterDisabled) Error() string {
	return fmt.Sprintf("El cluster no esta habilitado: %s", err.Name)
}
