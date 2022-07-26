package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
)

func main() {
	configSettings := config.LoadMainConfig()
	err := exec.Command(
		"rundll32",
		"url.dll,FileProtocolHandler",
		fmt.Sprintf("http://localhost:%s", configSettings.ServerPort),
	).Start()
	if err != nil {
		exception.ErrorLogger.Println(err)
	}
	fs := http.FileServer(http.Dir(configSettings.ResultUiRoot))
	http.Handle("/", fs)
	http.HandleFunc("/api/getmonthlycount", getMonthlyCountAPI)
	http.HandleFunc("/api/getskilltypelist", getSkillTypeListAPI)
	http.HandleFunc("/api/getskilltypebyid", getSkillTypeByIDAPI)
	http.HandleFunc("/api/getskilllist", getSkillListAPI)
	http.HandleFunc("/api/getskillbyid", getSkillByIDAPI)
	http.HandleFunc("/api/getallidandname", getAllSkillTypeAPI)
	fmt.Printf("Server listening on port %s\n", configSettings.ServerPort)
	exception.ErrorLogger.Panic(
		http.ListenAndServe(fmt.Sprintf(":%s", configSettings.ServerPort), nil),
	)
}

func makeResponse(w http.ResponseWriter, request *http.Request, content interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(content); err != nil {
		panic(err)
	}
}

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

func getSkillTypeListAPI(w http.ResponseWriter, request *http.Request) {
	skillTypeSlice, err := database.DbAdapter.SkillType.GetAllWithSkillNames()
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillTypeSlice)
}

func getSkillTypeByIDAPI(w http.ResponseWriter, request *http.Request) {
	skilTypeID, err := strconv.ParseUint(request.URL.Query().Get("skilltypeid"), 10, 64)
	fmt.Println(skilTypeID)
	if err != nil {
		panic(err)
	}
	skillType, err := database.DbAdapter.SkillType.GetByIDWithSkillNames(uint(skilTypeID))
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillType)
}

func getSkillListAPI(w http.ResponseWriter, request *http.Request) {
	skillSlice, err := database.DbAdapter.SkillName.GetAllWithTypeAndAliases()
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillSlice)
}

func getSkillByIDAPI(w http.ResponseWriter, request *http.Request) {
	skillID, err := strconv.ParseUint(request.URL.Query().Get("skillid"), 10, 64)
	if err != nil {
		panic(err)
	}
	skillName, err := database.DbAdapter.SkillName.GetByIDWithTypeAndAliases(uint(skillID))
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillName)
}

func getAllSkillTypeAPI(w http.ResponseWriter, request *http.Request) {
	skillTypeIDAndName, err := database.DbAdapter.SkillType.GetAllIDAndName()
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillTypeIDAndName)
}
