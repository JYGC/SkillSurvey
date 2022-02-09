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

func MakeSiteMap(sites []Site) (siteMap map[string]Site) {
	siteMap = make(map[string]Site)
	for _, site := range sites {
		siteMap[site.Name] = site
	}
	return siteMap
}
