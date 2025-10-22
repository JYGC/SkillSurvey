package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*MonthlyCountReport)(nil)

type MonthlyCountReport struct {
	core.BaseRecordProxy
}

func (r *MonthlyCountReport) Identifier() string {
	return r.GetString("identifier")
}
func (r *MonthlyCountReport) SetIdentifier(v string) {
	r.Set("identifier", v)
}

func (r *MonthlyCountReport) SkillName() string {
	return r.GetString("skillName")
}
func (r *MonthlyCountReport) SetSkillName(v string) {
	r.Set("skillName", v)
}

func (r *MonthlyCountReport) YearMonth() string {
	return r.GetString("YearMonth")
}
func (r *MonthlyCountReport) SetYearMonth(v string) {
	r.Set("YearMonth", v)
}

func (r *MonthlyCountReport) YearMonthDate() types.DateTime {
	return r.GetDateTime("yearMonthDate")
}
func (r *MonthlyCountReport) SetYearMonthDate(v types.DateTime) {
	r.Set("yearMonthDate", v)
}

func (r *MonthlyCountReport) Count() int {
	return r.GetInt("count")
}
func (r *MonthlyCountReport) SetCount(v int) {
	r.Set("count", v)
}
