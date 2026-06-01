package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*Role)(nil)

type Role struct {
	core.BaseRecordProxy
}

func (r *Role) Name() string        { return r.GetString("name") }
func (r *Role) Description() string { return r.GetString("description") }
