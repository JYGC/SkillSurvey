package database

import (
	"github.com/JYGC/SkillSurvey/internal/entities"
	"gorm.io/gorm"
)

type SkillNameTableCall struct {
	DbTableCallBase
}

func NewSkillNameTableCall(db *gorm.DB) (tableCall *SkillNameTableCall) {
	tableCall = new(SkillNameTableCall)
	tableCall.db = db
	tableCall.MigrateTable(&entities.SkillType{})
	tableCall.MigrateTable(&entities.SkillName{})
	tableCall.MigrateTable(&entities.SkillNameAlias{})
	return tableCall
}

func (s SkillNameTableCall) GetAlias() (result []entities.SkillNameAlias, err error) {
	err = s.db.Joins("SkillName").Find(&result).Error
	return result, err
}

// func (s SkillNameTableCall) GetAlias() (result []entities.AliasWithSkillName, err error) {
// 	err = s.db.Model(&entities.SkillName{}).Select(
// 		"skill_names.name, skill_name_aliases.alias",
// 	).Joins(
// 		"left join skill_name_aliases on skill_names.id = skill_name_aliases.skill_name_id",
// 	).Scan(&result).Error
// 	return result, err
// }

func (s SkillNameTableCall) GetByName(skillName string) (result entities.SkillName, err error) {
	err = s.db.Where("name = ?", skillName).First(&result).Error
	return result, err
}
