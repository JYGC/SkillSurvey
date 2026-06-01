package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*MonthlyCountReport)(nil)

type MonthlyCountReport struct {
	core.BaseRecordProxy
}

func (r *MonthlyCountReport) Identifier() string          { return r.GetString("identifier") }
func (r *MonthlyCountReport) SkillName() string            { return r.GetString("skillName") }
func (r *MonthlyCountReport) YearMonth() string            { return r.GetString("YearMonth") }
func (r *MonthlyCountReport) YearMonthDate() types.DateTime { return r.GetDateTime("yearMonthDate") }
func (r *MonthlyCountReport) Count() int                   { return r.GetInt("count") }
