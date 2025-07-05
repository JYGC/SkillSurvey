package main

import (
	"fmt"
	"net/http"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/handlers"
)

func main() {
	configSettings := config.LoadMainConfig()
	handlers.SetReportHandlers()
	handlers.SetSkillTypeHandlers()
	handlers.SetSkillHandlers()
	fmt.Printf("Server listening on port %s\n", configSettings.ServerPort)
	exception.ErrorLogger.Panic(
		http.ListenAndServe(fmt.Sprintf(":%s", configSettings.ServerPort), nil),
	)
}
