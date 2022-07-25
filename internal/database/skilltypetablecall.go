package database

import (
	"github.com/JYGC/SkillSurvey/internal/entities"
	"gorm.io/gorm"
)

type SkillTypeTableCall struct {
	DbTableCallBase
}

func NewSkillTypeTableCall(db *gorm.DB) (tableCall *SkillTypeTableCall) {
	tableCall = new(SkillTypeTableCall)
	tableCall.db = db
	tableCall.MigrateTable(&entities.SkillType{})
	return tableCall
}

func (s SkillTypeTableCall) GetAll() (skillTypeListResult []entities.SkillType, err error) {
	var skillTypeSlice []entities.SkillType
	err = s.db.Find(&skillTypeSlice).Error
	if err != nil {
		return nil, err
	}
	var skillNameSlice []entities.SkillName
	err = s.db.Model(&skillTypeSlice).Association("SkillNames").Find(&skillNameSlice)
	if err != nil {
		return nil, err
	}
	skillTypeIDMap := make(map[uint]entities.SkillType)
	for _, skillType := range skillTypeSlice {
		skillTypeIDMap[skillType.ID] = skillType
	}
	for _, skillName := range skillNameSlice {
		if skillType, ok := skillTypeIDMap[skillName.SkillTypeID]; ok {
			skillType.SkillNames = append(skillType.SkillNames, skillName)
			skillTypeIDMap[skillName.SkillTypeID] = skillType
		}
	}
	for _, skillType := range skillTypeIDMap {
		skillTypeListResult = append(skillTypeListResult, skillType)
	}
	return skillTypeListResult, err
}
