package entities

type SkillType struct {
	EntityBase
	Name string
	//Description string
	SkillNames []SkillName
}

type SkillName struct {
	EntityBase
	SkillTypeID      uint
	SkillType        SkillType `gorm:"foreignKey:SkillTypeID"`
	Name             string
	IsEnabled        bool
	SkillNameAliases []SkillNameAlias
}

type SkillNameAlias struct {
	EntityBase
	SkillNameID uint
	SkillName   SkillName `gorm:"foreignKey:SkillNameID"`
	Alias       string
}
