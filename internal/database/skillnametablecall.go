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
	if err = s.db.First(&skillNameResult.SkillType, skillNameResult.SkillTypeID).Error; err != nil {
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
		skillNameFromDB.SkillTypeID = changedSkillName.SkillTypeID
		if err = s.db.First(&skillNameFromDB.SkillType, skillNameFromDB.SkillTypeID).Error; err != nil {
			return err
		}
	}
	// modifying aliases
	changedOrNewAliasIDMap := make(map[uint]entities.SkillNameAlias)
	for _, alias := range changedSkillName.SkillNameAliases {
		changedOrNewAliasIDMap[alias.ID] = alias
	}
	// decrement because of deleting from skillNameFromDB.SkillNameAliases
	for index := len(skillNameFromDB.SkillNameAliases) - 1; index >= 0; index-- {
		alias := skillNameFromDB.SkillNameAliases[index]
		if changedAlias, ok := changedOrNewAliasIDMap[alias.ID]; ok {
			// update aliases
			skillNameFromDB.SkillNameAliases[index].Alias = changedAlias.Alias
			delete(changedOrNewAliasIDMap, alias.ID)
			continue
		}
		// remove deleted aliases
		s.db.Delete(&alias)
		skillNameFromDB.SkillNameAliases = skillNameFromDB.SkillNameAliases[:len(skillNameFromDB.SkillNameAliases)-1]
	}
	// add new aliases
	for _, alias := range changedOrNewAliasIDMap {
		skillNameFromDB.SkillNameAliases = append(skillNameFromDB.SkillNameAliases, alias)
	}
	skillNameFromDB.Name = changedSkillName.Name
	skillNameFromDB.IsEnabled = changedSkillName.IsEnabled
	if err = s.db.Save(&skillNameFromDB).Error; err != nil {
		return err
	}
	return err
}
