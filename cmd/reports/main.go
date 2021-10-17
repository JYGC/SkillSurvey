package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JYGC/SkillSurvey/internal/config"
)

func main() {
	configSettings := config.LoadMainConfig()
	fs := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fs)
	fmt.Printf("Server listening on port %s\n", configSettings.ServerPort)
	log.Panic(
		http.ListenAndServe(fmt.Sprintf(":%s", configSettings.ServerPort), nil),
	)
}
