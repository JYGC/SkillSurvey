package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*Site)(nil)

type Site struct {
	core.BaseRecordProxy
}

func (r *Site) Name() string                { return r.GetString("name") }
func (r *Site) SetName(v string)            { r.Set("name", v) }
func (r *Site) Created() types.DateTime     { return r.GetDateTime("created") }
func (r *Site) SetCreated(v types.DateTime) { r.Set("created", v) }
func (r *Site) Updated() types.DateTime     { return r.GetDateTime("updated") }
func (r *Site) SetUpdated(v types.DateTime) { r.Set("updated", v) }
