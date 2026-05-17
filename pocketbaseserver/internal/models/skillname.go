package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*SkillName)(nil)

type SkillName struct {
	core.BaseRecordProxy
}

func (r *SkillName) SkillType() string     { return r.GetString("skillType") }
func (r *SkillName) SetSkillType(v string) { r.Set("skillType", v) }
func (r *SkillName) Name() string          { return r.GetString("name") }
func (r *SkillName) SetName(v string)      { r.Set("name", v) }
func (r *SkillName) IsEnabled() bool       { return r.GetBool("isEnabled") }
func (r *SkillName) SetIsEnabled(v bool)   { r.Set("isEnabled", v) }
