package main

import (
	"fmt"
	"os"

	"github.com/JYGC/SkillSurvey/pkg/services"
)

func main() {
	var service services.IService = nil
	if len(os.Args) == 2 {
		selectedName := os.Args[1]
		switch selectedName {
		case "survey":
			service = services.NewSurvey()
		case "reports":
			service = services.NewReports()
		default:
			fmt.Println("Unknown argment")
		}

		if service != nil {
			service.Run()
		}
	} else {
		fmt.Println("Usage: ./skillsurvey <service name>")
	}
}
