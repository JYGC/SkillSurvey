package main

import (
	"strconv"
	"time"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/exception"
)

func main() {
	makeMonthlyCountReport()
}

func makeMonthlyCountReport() {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{"Variables": variableRef})
	// Associate skill name with array of their aliases
	skillNameAliases, err := database.DbAdapter.SkillName.GetAlias()
	if err != nil {
		panic(err)
	}
	skillNameMap := make(map[string][]string)
	for _, skillNameAlias := range skillNameAliases {
		skillNameMap[skillNameAlias.Name] = append(
			skillNameMap[skillNameAlias.Name],
			skillNameAlias.Alias,
		)
		// Skill names with no aliases must have an empty alias array
		if len(skillNameAlias.Alias) == 0 {
			skillNameMap[skillNameAlias.Name] = skillNameMap[skillNameAlias.Name][:len(skillNameMap[skillNameAlias.Name])-1]
		}
	}
	// Create report
	for skillName, aliases := range skillNameMap {
		skill, err := database.DbAdapter.SkillName.GetByName(skillName)
		if err != nil {
			variableRef["skillName"] = skillName
			panic(err)
		}
		counts, err := database.DbAdapter.JobPost.GetMonthlyCountBySkill(skillName, aliases)
		if err != nil {
			variableRef["skillName"] = skillName
			variableRef["aliases"] = aliases
			panic(err)
		}
		for i := range counts {
			counts[i].SkillName = skill
			counts[i].Identifier = strconv.FormatUint(uint64(skill.ID), 10) + " " + counts[i].YearMonth
			counts[i].YearMonthDate, err = time.Parse(time.RFC3339, counts[i].YearMonth+"-01T00:00:00Z")
			if err != nil {
				variableRef["i"] = i
				variableRef["skill.ID"] = skill.ID
				variableRef["counts[i].YearMonth"] = counts[i].YearMonth
				panic(err)
			}
		}
		if err := database.DbAdapter.MonthlyCount.BulkUpdateOrInsert(counts); err != nil {
			variableRef["counts"] = counts
			panic(err)
		}
	}
}
