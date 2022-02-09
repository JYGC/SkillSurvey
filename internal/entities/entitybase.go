package entities

type IEntity interface {
	ToInterface() map[string]interface{}
}

type EntityBase struct {
	ID uint
}
