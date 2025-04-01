package siteadapters

import (
	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

type ISiteAdapter interface {
	RunSurvey() []entities.InboundJobPost
}

type SiteAdapterBase struct {
	ConfigSettings config.SiteAdapterConfig
}
