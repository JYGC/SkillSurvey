package models

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ core.RecordProxy = (*JobPost)(nil)

type JobPost struct {
	core.BaseRecordProxy
}

func (r *JobPost) Site() string {
	return r.GetString("site")
}
func (r *JobPost) SetSite(v string) {
	r.Set("site", v)
}

func (r *JobPost) JobSiteNumber() string {
	return r.GetString("job_site_number")
}
func (r *JobPost) SetJobSiteNumber(v string) {
	r.Set("job_site_number", v)
}

func (r *JobPost) Title() string {
	return r.GetString("title")
}
func (r *JobPost) SetTitle(v string) {
	r.Set("title", v)
}

func (r *JobPost) Body() string {
	return r.GetString("body")
}
func (r *JobPost) SetBody(v string) {
	r.Set("body", v)
}

func (r *JobPost) PostedDate() types.DateTime {
	return r.GetDateTime("posted_date")
}
func (r *JobPost) SetPostedDate(v types.DateTime) {
	r.Set("posted_date", v)
}

func (r *JobPost) City() string {
	return r.GetString("city")
}
func (r *JobPost) SetCity(v string) {
	r.Set("city", v)
}

func (r *JobPost) Country() string {
	return r.GetString("country")
}
func (r *JobPost) SetCountry(v string) {
	r.Set("country", v)
}

func (r *JobPost) Suburb() string {
	return r.GetString("suburb")
}
func (r *JobPost) SetSuburb(v string) {
	r.Set("suburb", v)
}

func (r *JobPost) Created() types.DateTime {
	return r.GetDateTime("created")
}
func (r *JobPost) SetCreated(v types.DateTime) {
	r.Set("created", v)
}

func (r *JobPost) Updated() types.DateTime {
	return r.GetDateTime("updated")
}
func (r *JobPost) SetUpdated(v types.DateTime) {
	r.Set("updated", v)
}
