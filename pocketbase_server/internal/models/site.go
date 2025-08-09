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

func (s *Site) Url() string {
	return s.GetString("url")
}

func (s *Site) SetUrl(url string) {
	s.Set("url", url)
}
