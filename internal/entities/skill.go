package entities

type SkillType struct {
	EntityBase
	Name string
}

type SkillName struct {
	EntityBase
	SkillTypeID uint
	SkillType   SkillType `gorm:"foreignKey:SkillTypeID"`
	Name        string
	IsEnabled   bool
}

type SkillNameAlias struct {
	EntityBase
	SkillNameID uint
	SkillName   SkillName `gorm:"foreignKey:SkillNameID"`
	Alias       string
}

type AliasWithSkillName struct {
	Name  string
	Alias string
}
