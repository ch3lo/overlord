package report

import "fmt"

// BroadcasterAlreadyExist sucede cuando se esta agregando un broadcaster que ya existe
type BroadcasterAlreadyExist struct {
	Name string
}

func (err BroadcasterAlreadyExist) Error() string {
	return fmt.Sprintf("El broadcaster ya existe: %s", err.Name)
}
