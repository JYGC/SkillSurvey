package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*SkillName)(nil)

type SkillName struct {
	core.BaseRecordProxy
}

func (r *SkillName) SkillType() string     { return r.GetString("skill_type") }
func (r *SkillName) SetSkillType(v string) { r.Set("skill_type", v) }
func (r *SkillName) Name() string          { return r.GetString("name") }
func (r *SkillName) SetName(v string)      { r.Set("name", v) }
func (r *SkillName) IsEnabled() bool       { return r.GetBool("is_enabled") }
func (r *SkillName) SetIsEnabled(v bool)   { r.Set("is_enabled", v) }
