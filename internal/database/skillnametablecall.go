package database

import (
	"github.com/JYGC/SkillSurvey/internal/dataschemas"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"gorm.io/gorm"
)

type SkillNameTableCall struct {
	DbTableCallBase
}

func NewSkillNameTableCall(db *gorm.DB) (tableCall *SkillNameTableCall) {
	tableCall = new(SkillNameTableCall)
	tableCall.db = db
	tableCall.MigrateTable(&entities.SkillName{})
	tableCall.MigrateTable(&entities.SkillNameAlias{})
	return tableCall
}

func (s SkillNameTableCall) GetAlias() (result []dataschemas.AliasWithSkillName, err error) {
	err = s.db.Model(&entities.SkillName{}).Select(
		"skill_names.name, skill_name_aliases.alias",
	).Joins(
		"left join skill_name_aliases on skill_names.id = skill_name_aliases.skill_name_id",
	).Scan(&result).Error
	return result, err
}

func (s SkillNameTableCall) GetByName(skillName string) (result entities.SkillName, err error) {
	err = s.db.Where("name = ?", skillName).First(&result).Error
	return result, err
}

func (s SkillNameTableCall) GetAll() (skillNameListResult []entities.SkillName, err error) {
	var skillNameSlice []entities.SkillName
	err = s.db.Find(&skillNameSlice).Error
	if err != nil {
		return nil, err
	}
	var skillTypeSlice []entities.SkillType
	err = s.db.Find(&skillTypeSlice).Error
	if err != nil {
		return nil, err
	}
	skillTypeIDMap := make(map[uint]entities.SkillType)
	for _, skillType := range skillTypeSlice {
		skillTypeIDMap[skillType.ID] = skillType
	}
	skillNameIDMap := make(map[uint]entities.SkillName)
	for _, skillName := range skillNameSlice {
		skillName.SkillType = skillTypeIDMap[skillName.SkillTypeID]
		skillNameIDMap[skillName.ID] = skillName
	}
	// get aliases and attach them to SkillNames
	var skillNameAliasSlice []entities.SkillNameAlias
	err = s.db.Model(&skillNameSlice).Association("SkillNameAliases").Find(&skillNameAliasSlice)
	if err != nil {
		return nil, err
	}
	for _, skillNameAlias := range skillNameAliasSlice {
		if skillName, ok := skillNameIDMap[skillNameAlias.SkillNameID]; ok {
			skillName.SkillNameAliases = append(skillName.SkillNameAliases, skillNameAlias)
			skillNameIDMap[skillNameAlias.SkillNameID] = skillName
		}
	}
	for _, skillName := range skillNameIDMap {
		skillNameListResult = append(skillNameListResult, skillName)
	}
	return skillNameListResult, err
}

func (s SkillNameTableCall) GetByID(ID uint) (skillNameResult *entities.SkillName, err error) {
	err = s.db.First(&skillNameResult, ID).Error
	if err != nil {
		return nil, err
	}
	err = s.db.First(&skillNameResult.SkillType, skillNameResult.SkillTypeID).Error
	if err != nil {
		return nil, err
	}
	err = s.db.Model(&skillNameResult).Association("SkillNameAliases").Find(&skillNameResult.SkillNameAliases)
	if err != nil {
		return nil, err
	}
	return skillNameResult, err
}
