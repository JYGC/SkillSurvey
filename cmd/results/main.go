package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

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
	fs := http.FileServer(http.Dir("./resultsui/dist"))
	http.Handle("/", fs)
	http.HandleFunc("/api/", getMonthlyCount)
	fmt.Printf("Server listening on port %s\n", configSettings.ServerPort)
	exception.ErrorLogger.Panic(
		http.ListenAndServe(fmt.Sprintf(":%s", configSettings.ServerPort), nil),
	)
}

func getMonthlyCount(w http.ResponseWriter, request *http.Request) {
	//decoder := json.NewDecoder(request.Body)
	//var _inp inp = inp{Num1: 3, Num2: 2}
	//var _resp resp
	//decoder.Decode(&_inp)
	//_resp.Num = _inp.Num1 + _inp.Num2
	//fmt.Println(_resp)

	reportSlice, err := database.DbAdapter.MonthlyCount.GetReport()
	_resp := make(map[string][]entities.MonthlyCountReport)
	for _, reportElement := range reportSlice {
		_resp[reportElement.SkillName.Name] = append(_resp[reportElement.SkillName.Name], reportElement)
	}
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(_resp); err != nil {
		panic(err)
	}
}
