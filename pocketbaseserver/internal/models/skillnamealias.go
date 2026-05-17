package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*SkillNameAlias)(nil)

type SkillNameAlias struct {
	core.BaseRecordProxy
}

func (r *SkillNameAlias) Alias() string               { return r.GetString("alias") }
func (r *SkillNameAlias) SetAlias(v string)           { r.Set("alias", v) }
func (r *SkillNameAlias) SkillName() string           { return r.GetString("skillName") }
func (r *SkillNameAlias) SetSkillName(v string)       { r.Set("skillName", v) }
func (r *SkillNameAlias) Created() types.DateTime     { return r.GetDateTime("created") }
func (r *SkillNameAlias) SetCreated(v types.DateTime) { r.Set("created", v) }
func (r *SkillNameAlias) Updated() types.DateTime     { return r.GetDateTime("updated") }
func (r *SkillNameAlias) SetUpdated(v types.DateTime) { r.Set("updated", v) }
