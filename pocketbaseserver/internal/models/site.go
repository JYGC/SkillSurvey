package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*Site)(nil)

type Site struct {
	core.BaseRecordProxy
}

func (r *Site) Name() string            { return r.GetString("name") }
func (r *Site) Url() string             { return r.GetString("url") }
func (r *Site) Created() types.DateTime { return r.GetDateTime("created") }
func (r *Site) Updated() types.DateTime { return r.GetDateTime("updated") }
