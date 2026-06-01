package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*SkillNameAlias)(nil)

type SkillNameAlias struct {
	core.BaseRecordProxy
}

func (r *SkillNameAlias) Alias() string           { return r.GetString("alias") }
func (r *SkillNameAlias) SkillName() string       { return r.GetString("skillName") }
func (r *SkillNameAlias) Created() types.DateTime { return r.GetDateTime("created") }
func (r *SkillNameAlias) Updated() types.DateTime { return r.GetDateTime("updated") }
