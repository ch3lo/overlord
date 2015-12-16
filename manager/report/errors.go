package report

import "fmt"

// BroadcastWorkerAlreadyExist sucede cuando se esta agregando un broadcast worker que ya existe
type BroadcastWorkerAlreadyExist struct {
	Name string
}

func (err BroadcastWorkerAlreadyExist) Error() string {
	return fmt.Sprintf("El broadcast worker ya existe: %s", err.Name)
}
