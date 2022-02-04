package main

import (
	"fmt"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

func main() {
	// configSettings := config.LoadMainConfig()
	// err := exec.Command(
	// 	"rundll32",
	// 	"url.dll,FileProtocolHandler",
	// 	fmt.Sprintf("http://localhost:%s", configSettings.ServerPort),
	// ).Start()
	// if err != nil {
	// 	exception.ErrorLogger.Println(err)
	// }
	// fs := http.FileServer(http.Dir("./frontend/dist"))
	// http.Handle("/", fs)
	// fmt.Printf("Server listening on port %s\n", configSettings.ServerPort)
	// exception.ErrorLogger.Panic(
	// 	http.ListenAndServe(fmt.Sprintf(":%s", configSettings.ServerPort), nil),
	// )
	//////////////////////////
	skillNameAliases, err := database.DbAdapter.SkillName.GetAlias()
	if err != nil {
		fmt.Println(err)
	}
	skillNameMap := make(map[string][]entities.SkillNameAlias)
	for _, skillNameAlias := range skillNameAliases {
		skillNameMap[skillNameAlias.SkillName.Name] = append(
			skillNameMap[skillNameAlias.SkillName.Name],
			skillNameAlias,
		)
	}
	for k, v := range skillNameMap {
		r, e := database.DbAdapter.JobPost.GetMonthlyCountBySkill(k, v)
		fmt.Println(r, e)
	}
}
