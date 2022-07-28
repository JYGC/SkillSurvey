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

func (s SkillTypeTableCall) GetAllWithSkillNames() (skillTypeListResult []entities.SkillType, err error) {
	var skillTypeSlice []entities.SkillType
	if err = s.db.Find(&skillTypeSlice).Error; err != nil {
		return nil, err
	}
	skillTypeIDMap := make(map[uint]entities.SkillType)
	for _, skillType := range skillTypeSlice {
		skillTypeIDMap[skillType.ID] = skillType
	}
	var skillNameSlice []entities.SkillName
	if err = s.db.Model(&skillTypeSlice).Association("SkillNames").Find(&skillNameSlice); err != nil {
		return nil, err
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

func (s SkillTypeTableCall) GetByIDWithSkillNames(ID uint) (skillTypeResult *entities.SkillType, err error) {
	if err = s.db.First(&skillTypeResult, ID).Error; err != nil {
		return nil, err
	}
	if err = s.db.Model(&skillTypeResult).Association("SkillNames").Find(&skillTypeResult.SkillNames); err != nil {
		return nil, err
	}
	return skillTypeResult, err
}

func (s SkillTypeTableCall) GetAllIDAndName() (skillTypeMapResult map[uint]string, err error) {
	skillTypeMapResult = make(map[uint]string)
	var skillTypeSlice []entities.SkillType
	if err = s.db.Find(&skillTypeSlice).Error; err != nil {
		return nil, err
	}
	for _, skillType := range skillTypeSlice {
		skillTypeMapResult[skillType.ID] = skillType.Name
	}
	return skillTypeMapResult, err
}
