package siteadapters

import (
	"github.com/JYGC/SkillSurvey/internal/entities"
)

type ISiteAdapter interface {
	RunSurvey() []entities.InboundJobPost
}
