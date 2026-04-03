package models

import (
	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*Site)(nil)

type UserRole struct {
	core.BaseRecordProxy
}

func (r *SkillType) User() string     { return r.GetString("user") }
func (r *SkillType) SetUser(v string) { r.Set("user", v) }
func (r *SkillType) Role() string     { return r.GetString("role") }
func (r *SkillType) SetRole(v string) { r.Set("role", v) }
