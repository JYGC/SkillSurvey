package handlers

import (
	"net/http"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

func getMonthlyCountAPI(w http.ResponseWriter, request *http.Request) {
	reportSlice, err := database.DbAdapter.MonthlyCount.GetReport()
	report := make(map[string][]entities.MonthlyCountReport)
	for _, reportElement := range reportSlice {
		report[reportElement.SkillName.Name] = append(report[reportElement.SkillName.Name], reportElement)
	}
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, report)
}

func SetReportHandlers() {
	http.HandleFunc("/report/getmonthlycount", getMonthlyCountAPI)
}
