package models

import "github.com/pocketbase/pocketbase/core"

var _ core.RecordProxy = (*Site)(nil)

type Site struct {
	core.BaseRecordProxy
	//Name string
}

func (s *Site) Name() string {
	return s.GetString("name")
}

func (s *Site) SetName(name string) {
	s.Set("name", name)
}

func (s *Site) BaseUrl() string {
	return s.GetString("base_url")
}

func (s *Site) SetBaseUrl(baseUrl string) {
	s.Set("base_url", baseUrl)
}
