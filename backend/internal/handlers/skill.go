package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

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

func addSkillAPI(responseWriter http.ResponseWriter, request *http.Request) {
	var skillNameID uint
	var requestBody map[string]interface{}
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		panic(err)
	}
	newSkillName := entities.SkillName{}
	var err error
	var ok bool
	var uint64SkillTypeID uint64
	if uint64SkillTypeID, err = strconv.ParseUint(fmt.Sprintf("%v", requestBody["SkillTypeID"]), 10, 64); err != nil {
		panic(err)
	}
	newSkillName.SkillTypeID = uint(uint64SkillTypeID)
	if newSkillName.Name, ok = requestBody["Name"].(string); !ok {
		panic("can't convert Name to string")
	}
	newSkillName.IsEnabled = true
	if reflect.TypeOf(requestBody["SkillNameAliases"]).Kind() != reflect.Slice {
		panic("can't convert SkillNameAliases to slice")
	}
	skillNameAliasesValue := reflect.ValueOf(requestBody["SkillNameAliases"])
	for index := 0; index < skillNameAliasesValue.Len(); index++ {
		skillNameAliasInterface := skillNameAliasesValue.Index(index).Interface()
		skillNameAliasMap := skillNameAliasInterface.(map[string]interface{})
		newSkillName.SkillNameAliases = append(newSkillName.SkillNameAliases, entities.SkillNameAlias{
			Alias: skillNameAliasMap["Alias"].(string),
		})
	}
	if err = getErrorIfSkillOrItsAliasesNotValid(newSkillName); err != nil {
		panic(err)
	}
	if skillNameID, err = database.DbAdapter.SkillName.AddOne(newSkillName); err != nil {
		panic(err)
	}
	makeResponse(responseWriter, request, map[string]interface{}{"ID": skillNameID})
}

func saveSkillAPI(responseWriter http.ResponseWriter, request *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		panic(err)
	}
	var changedSkillName entities.SkillName = entities.SkillName{}
	var err error
	var ok bool
	var uint64SkillNameID uint64
	if uint64SkillNameID, err = strconv.ParseUint(fmt.Sprintf("%v", requestBody["ID"]), 10, 64); err != nil {
		panic(err)
	}
	changedSkillName.ID = uint(uint64SkillNameID)
	var uint64SkillTypeID uint64
	if uint64SkillTypeID, err = strconv.ParseUint(fmt.Sprintf("%v", requestBody["SkillTypeID"]), 10, 64); err != nil {
		panic(err)
	}
	changedSkillName.SkillTypeID = uint(uint64SkillTypeID)
	if changedSkillName.Name, ok = requestBody["Name"].(string); !ok {
		panic("can't convert Name to string")
	}
	changedSkillName.IsEnabled = true
	if reflect.TypeOf(requestBody["SkillNameAliases"]).Kind() != reflect.Slice {
		panic("can't convert SkillNameAliases to slice")
	}
	skillNameAliasesValue := reflect.ValueOf(requestBody["SkillNameAliases"])
	for index := 0; index < skillNameAliasesValue.Len(); index++ {
		skillNameAliasInterface := skillNameAliasesValue.Index(index).Interface()
		skillNameAliasMap := skillNameAliasInterface.(map[string]interface{})
		changedSkillName.SkillNameAliases = append(changedSkillName.SkillNameAliases, entities.SkillNameAlias{
			Alias: skillNameAliasMap["Alias"].(string),
		})
	}
	if err = getErrorIfSkillOrItsAliasesNotValid(changedSkillName); err != nil {
		panic(err)
	}
	if err = database.DbAdapter.SkillName.SaveOneWithTypeAndAliases(changedSkillName); err != nil {
		panic(err)
	}
	makeResponse(responseWriter, request, "success")
}

func getErrorIfSkillOrItsAliasesNotValid(skillName entities.SkillName) (err error) {
	if strings.TrimSpace(skillName.Name) == "" {
		return errors.New("skill must have name")
	}
	if skillName.SkillTypeID == 0 {
		return errors.New("skill must have type")
	}
	for _, alias := range skillName.SkillNameAliases {
		if strings.TrimSpace(alias.Alias) == "" {
			return errors.New("can't allow empty aliases")
		}
	}
	return nil
}

func deleteSkillAPI(responseWriter http.ResponseWriter, request *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		panic(err)
	}
	var err error
	var uint64SkillNameID uint64
	if uint64SkillNameID, err = strconv.ParseUint(fmt.Sprintf("%v", requestBody["ID"]), 10, 64); err != nil {
		panic(err)
	}
	if err = database.DbAdapter.SkillName.DeleteOne(uint(uint64SkillNameID)); err != nil {
		panic(err)
	}
	makeResponse(responseWriter, request, "success")
}

func SetSkillHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/skill/getall", getSkillListAPI)
	mux.HandleFunc("/skill/getbyid", getSkillByIDAPI)
	mux.HandleFunc("/skill/add", addSkillAPI)
	mux.HandleFunc("/skill/save", saveSkillAPI)
	mux.HandleFunc("/skill/delete", deleteSkillAPI)
}
