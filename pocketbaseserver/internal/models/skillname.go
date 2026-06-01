package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*SkillName)(nil)

type SkillName struct {
	core.BaseRecordProxy
}

func (r *SkillName) SkillType() string { return r.GetString("skillType") }
func (r *SkillName) Name() string      { return r.GetString("name") }
func (r *SkillName) IsEnabled() bool   { return r.GetBool("isEnabled") }
