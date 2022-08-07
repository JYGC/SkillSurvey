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

func (s SkillNameTableCall) GetAliasWithSkillName() (result []dataschemas.AliasWithSkillName, err error) {
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

func (s SkillNameTableCall) GetAllWithTypeAndAliases() (skillNameListResult []entities.SkillName, err error) {
	var skillNameSlice []entities.SkillName
	if err = s.db.Find(&skillNameSlice).Error; err != nil {
		return nil, err
	}
	var skillTypeSlice []entities.SkillType
	if err = s.db.Find(&skillTypeSlice).Error; err != nil {
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
	if err = s.db.Model(&skillNameSlice).Association("SkillNameAliases").Find(&skillNameAliasSlice); err != nil {
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

func (s SkillNameTableCall) GetByIDWithTypeAndAliases(ID uint) (skillNameResult *entities.SkillName, err error) {
	if err = s.db.First(&skillNameResult, ID).Error; err != nil {
		return nil, err
	}
	if err = s.db.Model(&skillNameResult).Association("SkillType").Find(&skillNameResult.SkillType); err != nil {
		return nil, err
	}
	if err = s.db.Model(&skillNameResult).Association("SkillNameAliases").Find(&skillNameResult.SkillNameAliases); err != nil {
		return nil, err
	}
	return skillNameResult, err
}

func (s SkillNameTableCall) AddOne(skillName entities.SkillName) (skillNameID uint, err error) {
	if err = s.db.Create(&skillName).Error; err != nil {
		return 0, err
	}
	return skillName.ID, err
}

func (s SkillNameTableCall) SaveOneWithTypeAndAliases(changedSkillName entities.SkillName) (err error) {
	var skillNameFromDB *entities.SkillName
	if skillNameFromDB, err = s.GetByIDWithTypeAndAliases(changedSkillName.ID); err != nil {
		return err
	}
	if skillNameFromDB.SkillTypeID != changedSkillName.SkillTypeID {
		var switchToSkillTypeFromDB *entities.SkillType
		if err = s.db.First(&switchToSkillTypeFromDB, changedSkillName.SkillTypeID).Error; err != nil {
			return err
		}
		if err = s.db.Model(&skillNameFromDB).Association("SkillType").Replace(switchToSkillTypeFromDB, &skillNameFromDB.SkillType); err != nil {
			return err
		}
	}
	// modifying aliases
	aliasFromDBIDMap := make(map[uint]entities.SkillNameAlias)
	for _, alias := range skillNameFromDB.SkillNameAliases {
		aliasFromDBIDMap[alias.ID] = alias
	}
	for index := range changedSkillName.SkillNameAliases {
		alias := changedSkillName.SkillNameAliases[index]
		if _, ok := aliasFromDBIDMap[alias.ID]; !ok {
			// New alias
			if err = s.db.Model(&skillNameFromDB).Association("SkillNameAliases").Append(&alias); err != nil {
				return err
			}
			continue
		}
		if err = s.db.Save(&alias).Error; err != nil {
			return err
		}
		delete(aliasFromDBIDMap, alias.ID)
	}
	// remove deleted aliases if any
	if len(aliasFromDBIDMap) > 0 {
		var aliasesToDelete []entities.SkillNameAlias
		var aliasIDsToDelete []uint
		for _, aliasToDelete := range aliasFromDBIDMap {
			aliasesToDelete = append(aliasesToDelete, aliasToDelete)
			aliasIDsToDelete = append(aliasIDsToDelete, aliasToDelete.ID)
		}
		if err = s.db.Model(&skillNameFromDB).Association("SkillNameAliases").Delete(aliasesToDelete); err != nil {
			return err
		}
		if err = s.db.Delete(&aliasesToDelete, aliasIDsToDelete).Error; err != nil {
			return err
		}
	}
	// change other skillname details
	skillNameFromDB.Name = changedSkillName.Name
	skillNameFromDB.IsEnabled = changedSkillName.IsEnabled
	if err = s.db.Save(&skillNameFromDB).Error; err != nil {
		return err
	}
	return err
}
