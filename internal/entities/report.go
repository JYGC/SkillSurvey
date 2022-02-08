package entities

import "time"

type MonthlyCountReport struct {
	EntityBase
	Identifier    string
	SkillNameID   uint
	SkillName     SkillName `gorm:"foreignKey:SkillNameID"`
	YearMonth     string    `gorm:"column:[YearMonth]"`
	YearMonthDate time.Time
	Count         int
}
