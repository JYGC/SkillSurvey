package database

import (
	"fmt"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/JYGC/SkillSurvey/pkg/entities"
)

type SiteCollection struct {
	CollectionBase
}

func NewSiteCollection(database *db.DB) *SiteCollection {
	collection := new(SiteCollection)
	collection.database = database
	collection.CollectionName = "Sites"
	existingCollection := collection.database.ColExists(collection.CollectionName)
	if !existingCollection {
		if err := collection.database.Create(collection.CollectionName); err != nil {
			panic(err)
		}
	}
	collection.collection = collection.database.Use(collection.CollectionName)
	if !existingCollection {
		//collection.collection.
		//TODO: tidot or GORM???
	}
	return collection
}

func (s SiteCollection) GetAll() {
	s.collection.ForEachDoc(func(id int, docCont []byte) bool {
		fmt.Println("Document", id, "is", string(docCont))
		return true
	})
}

func (s SiteCollection) InsertBulk() {
	siteSlice := make([]*entities.Site, 2)
	siteSlice[0] = new(entities.Site)
	siteSlice[0].Name = "https://www.seek.com.au"
	siteSlice[1] = new(entities.Site)
	siteSlice[1].Name = "https://au.jora.com"
	for _, site := range siteSlice {
		s.collection.Insert(site.ToInterface())
	}
}
