package database

import (
	"github.com/JYGC/SkillSurvey/internal/entities"
	"gorm.io/gorm"
)

type MonthlyCountReportTableCall struct {
	DbTableCallBase
}

func NewMonthlyCountReportTableCall(db *gorm.DB) (tableCall *MonthlyCountReportTableCall) {
	tableCall = new(MonthlyCountReportTableCall)
	tableCall.db = db
	tableCall.MigrateTable(&entities.MonthlyCountReport{})
	return tableCall
}

func (m MonthlyCountReportTableCall) BulkUpdateOrInsert(
	monthlyCountSlice []entities.MonthlyCountReport,
) (err error) {
	mounthlyCountReportMap := make(map[string]entities.MonthlyCountReport)
	var monthlyCountIdentifiers []string
	for _, monthlyCount := range monthlyCountSlice {
		monthlyCountIdentifiers = append(monthlyCountIdentifiers, monthlyCount.Identifier)
		mounthlyCountReportMap[monthlyCount.Identifier] = monthlyCount
	}
	// update exising mountlycountreports
	var existingMonthlyCountSlice []entities.MonthlyCountReport
	err = m.db.Where("identifier IN ?", monthlyCountIdentifiers).Find(&existingMonthlyCountSlice).Error
	if err != nil {
		return err
	}
	for _, existingMonthlyCount := range existingMonthlyCountSlice {
		existingMonthlyCount.Count = mounthlyCountReportMap[existingMonthlyCount.Identifier].Count
		delete(mounthlyCountReportMap, existingMonthlyCount.Identifier)
		m.db.Save(&existingMonthlyCount)
	}
	if len(mounthlyCountReportMap) == 0 {
		return err
	}
	// insert new monthly count
	var newMonthlyCount []entities.MonthlyCountReport
	for i := range mounthlyCountReportMap {
		newMonthlyCount = append(newMonthlyCount, mounthlyCountReportMap[i])
	}
	err = m.db.Create(&newMonthlyCount).Error
	return err
}
