package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*SkillType)(nil)

type SkillType struct {
	core.BaseRecordProxy
}

func (r *SkillType) Name() string            { return r.GetString("name") }
func (r *SkillType) SetName(v string)        { r.Set("name", v) }
func (r *SkillType) Description() string     { return r.GetString("description") }
func (r *SkillType) SetDescription(v string) { r.Set("description", v) }
