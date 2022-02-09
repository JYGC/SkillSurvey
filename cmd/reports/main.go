package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/JYGC/SkillSurvey/internal/database"
)

func main() {
	makeMonthlyCountReport()
}

func makeMonthlyCountReport() {
	skillNameAliases, err := database.DbAdapter.SkillName.GetAlias()
	if err != nil {
		fmt.Println(err)
	}
	// Associate skill name with array of their aliases
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
	for skillName, aliases := range skillNameMap {
		skill, e0 := database.DbAdapter.SkillName.GetByName(skillName)
		counts, e1 := database.DbAdapter.JobPost.GetMonthlyCountBySkill(skillName, aliases)
		if e0 != nil || e1 != nil {
			fmt.Println(e0)
			fmt.Println(e1)
		}
		for i := range counts {
			counts[i].SkillName = skill
			counts[i].Identifier = strconv.FormatUint(uint64(skill.ID), 10) + " " + counts[i].YearMonth
			var ec error
			counts[i].YearMonthDate, ec = time.Parse(time.RFC3339, counts[i].YearMonth+"-01T00:00:00Z")
			if ec != nil {
				fmt.Println(ec)
			}
		}
		database.DbAdapter.MonthlyCount.BulkUpdateOrInsert(counts)
	}
}
