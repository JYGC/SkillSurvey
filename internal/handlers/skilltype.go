package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

func getSkillTypeListAPI(w http.ResponseWriter, request *http.Request) {
	skillTypeSlice, err := database.DbAdapter.SkillType.GetAllWithSkillNames()
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillTypeSlice)
}

func getSkillTypeByIDAPI(w http.ResponseWriter, request *http.Request) {
	skilTypeID, err := strconv.ParseUint(request.URL.Query().Get("skilltypeid"), 10, 64)
	if err != nil {
		panic(err)
	}
	skillType, err := database.DbAdapter.SkillType.GetByIDWithSkillNames(uint(skilTypeID))
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillType)
}

func getAllSkillTypeIDAndNameAPI(w http.ResponseWriter, request *http.Request) {
	skillTypeIDAndName, err := database.DbAdapter.SkillType.GetAllIDAndName()
	if err != nil {
		panic(err)
	}
	makeResponse(w, request, skillTypeIDAndName)
}

func addSkillTypeAPI(responseWriter http.ResponseWriter, request *http.Request) {
	var skillTypeID uint
	var err error
	var requestBody map[string]interface{}
	if err = json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		panic(err)
	}
	newSkillType := entities.SkillType{}
	var ok bool
	if newSkillType.Name, ok = requestBody["Name"].(string); !ok {
		panic("can't convert Name to string")
	}
	if newSkillType.Description, ok = requestBody["Description"].(string); !ok {
		panic("can't convert Description to string")
	}
	if skillTypeID, err = database.DbAdapter.SkillType.AddOne(newSkillType); err != nil {
		panic(err)
	}
	makeResponse(responseWriter, request, map[string]interface{}{"ID": skillTypeID})
}

func saveSkillTypeAPI(responseWriter http.ResponseWriter, request *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		panic(err)
	}
	var changedSkillType entities.SkillType = entities.SkillType{}
	var uint64SkillTypeID uint64
	var err error
	var ok bool
	if uint64SkillTypeID, err = strconv.ParseUint(fmt.Sprintf("%v", requestBody["ID"]), 10, 64); err != nil {
		panic(err)
	}
	changedSkillType.ID = uint(uint64SkillTypeID)
	if changedSkillType.Name, ok = requestBody["Name"].(string); !ok {
		panic("can't convert Name to string")
	}
	if changedSkillType.Description, ok = requestBody["Description"].(string); !ok {
		panic("can't convert Description to string")
	}
	if err = database.DbAdapter.SkillType.SaveOne(changedSkillType); err != nil {
		panic(err)
	}
	makeResponse(responseWriter, request, "success")
}

func deleteSkillTypeAPI(responseWriter http.ResponseWriter, request *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		panic(err)
	}
	var err error
	var uint64SkillTypeID uint64
	if uint64SkillTypeID, err = strconv.ParseUint(fmt.Sprintf("%v", requestBody["ID"]), 10, 64); err != nil {
		panic(err)
	}
	if err = database.DbAdapter.SkillType.DeleteOne(uint(uint64SkillTypeID)); err != nil {
		panic(err)
	}
	makeResponse(responseWriter, request, "success")
}

func SetSkillTypeHandlers() {
	http.HandleFunc("/skilltype/getall", getSkillTypeListAPI)
	http.HandleFunc("/skilltype/getbyid", getSkillTypeByIDAPI)
	http.HandleFunc("/skilltype/getallidandname", getAllSkillTypeIDAndNameAPI)
	http.HandleFunc("/skilltype/add", addSkillTypeAPI)
	http.HandleFunc("/skilltype/save", saveSkillTypeAPI)
	http.HandleFunc("/skilltype/delete", deleteSkillTypeAPI)
}
