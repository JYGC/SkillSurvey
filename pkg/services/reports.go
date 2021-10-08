package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JYGC/SkillSurvey/pkg/config"
)

type Reports struct {
	ServiceBase
}

func NewReports() *Reports {
	reports := new(Reports)
	return reports
}

func (r *Reports) Run() {
	configSettings := config.LoadMainConfig()
	fs := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fs)
	fmt.Printf("Server listening on port %s\n", configSettings.ServerPort)
	log.Panic(
		http.ListenAndServe(fmt.Sprintf(":%s", configSettings.ServerPort), nil),
	)
}
