package api

import "fmt"

type CustomStatusAndMessageError interface {
	error
	GetStatus() int
	GetMessage() string
}

type ElementAlreadyExists struct {
}

func (e *ElementAlreadyExists) GetStatus() int {
	return 400
}

func (e *ElementAlreadyExists) GetMessage() string {
	return "Elemento ya existe"
}

func (ce *ElementAlreadyExists) Error() string {
	return fmt.Sprintf("%v: %v", ce.GetStatus(), ce.GetMessage())
}

type ServiceNotFound struct {
}

func (e *ServiceNotFound) GetStatus() int {
	return 400
}

func (e *ServiceNotFound) GetMessage() string {
	return "Servicio no existe"
}

func (ce *ServiceNotFound) Error() string {
	return fmt.Sprintf("%v: %v", ce.GetStatus(), ce.GetMessage())
}

type SerializationError struct {
	Message string
}

func (e *SerializationError) GetStatus() int {
	return 400
}

func (e *SerializationError) GetMessage() string {
	return e.Message
}

func (ce *SerializationError) Error() string {
	return fmt.Sprintf("%v: %v", ce.GetStatus(), ce.GetMessage())
}

type UnknownError struct {
}

func (e *UnknownError) GetStatus() int {
	return 500
}

func (e *UnknownError) GetMessage() string {
	return "Error desconocido"
}

func (ce *UnknownError) Error() string {
	return fmt.Sprintf("%v: %v", ce.GetStatus(), ce.GetMessage())
}

type ImageNameRegexpError struct {
	Message string
}

func (e *ImageNameRegexpError) GetStatus() int {
	return 400
}

func (e *ImageNameRegexpError) GetMessage() string {
	return e.Message
}

func (ce *ImageNameRegexpError) Error() string {
	return fmt.Sprintf("%v: %v", ce.GetStatus(), ce.GetMessage())
}

// NO USADO AUN
type ServiceVersionNotFound struct {
}

func (e *ServiceVersionNotFound) GetStatus() int {
	return 400
}

func (e *ServiceVersionNotFound) GetMessage() string {
	return "Version del servicio no existe"
}

func (ce *ServiceVersionNotFound) Error() string {
	return fmt.Sprintf("%v: %v", ce.GetStatus(), ce.GetMessage())
}
