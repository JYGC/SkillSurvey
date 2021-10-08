package entities

type Site struct {
	EntityBase
	Name string
}

func (s Site) ToInterface() map[string]interface{} {
	return map[string]interface{}{
		"Name": s.Name,
	}
}
