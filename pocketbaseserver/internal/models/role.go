package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*Site)(nil)

type Role struct {
	core.BaseRecordProxy
}

func (r *Site) Name() string     { return r.GetString("name") }
func (r *Site) SetName(v string) { r.Set("name", v) }

func (r *Site) Description() string     { return r.GetString("description") }
func (r *Site) SetDescription(v string) { r.Set("description", v) }
