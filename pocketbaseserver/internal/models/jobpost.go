package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*JobPost)(nil)

type JobPost struct {
	core.BaseRecordProxy
}

func (r *JobPost) Site() string             { return r.GetString("site") }
func (r *JobPost) JobSiteNumber() string    { return r.GetString("jobSiteNumber") }
func (r *JobPost) Title() string            { return r.GetString("title") }
func (r *JobPost) Body() string             { return r.GetString("body") }
func (r *JobPost) PostedDate() types.DateTime { return r.GetDateTime("postedDate") }
func (r *JobPost) City() string             { return r.GetString("city") }
func (r *JobPost) Country() string          { return r.GetString("country") }
func (r *JobPost) Suburb() string           { return r.GetString("suburb") }
func (r *JobPost) Created() types.DateTime  { return r.GetDateTime("created") }
func (r *JobPost) Updated() types.DateTime  { return r.GetDateTime("updated") }
