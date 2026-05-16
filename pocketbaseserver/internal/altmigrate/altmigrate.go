package altmigrate

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// Summary holds the result counts from a single migration run.
type Summary struct {
	Attempted int
	Written   int
	Skipped   int // record already existed in PocketBase
	Failed    int
}

type legacyJobPost struct {
	ID            uint
	SiteID        uint
	JobSiteNumber string
	Title         string
	Body          string
	PostedDate    string
	City          string
	Country       string
	Suburb        string
}

// Run reads jobPosts from the legacy SQLite database at legacyDbPath and writes
// them into PocketBase via the internal app API, bypassing HTTP.
// Prerequisite: run the existing migrate tool first so that legacy site names
// (e.g. "seek.com.au") already exist as PocketBase site records.
// Individual record failures are logged and counted but do not abort the run.
func Run(app core.App, legacyDbPath string) (Summary, error) {
	db, err := sql.Open("sqlite", legacyDbPath)
	if err != nil {
		return Summary{}, fmt.Errorf("open legacy db: %w", err)
	}
	defer db.Close()

	siteMap, err := buildSiteMap(app, db)
	if err != nil {
		return Summary{}, err
	}

	col, err := app.FindCollectionByNameOrId("jobPosts")
	if err != nil {
		return Summary{}, fmt.Errorf("find jobPosts collection: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, site_id, job_site_number, title, body,
		       posted_date, city, country, suburb
		FROM   job_posts
	`)
	if err != nil {
		return Summary{}, fmt.Errorf("query job_posts: %w", err)
	}
	defer rows.Close()

	var s Summary
	for rows.Next() {
		var jp legacyJobPost
		if err := rows.Scan(
			&jp.ID, &jp.SiteID, &jp.JobSiteNumber, &jp.Title, &jp.Body,
			&jp.PostedDate, &jp.City, &jp.Country, &jp.Suburb,
		); err != nil {
			log.Printf("altmigrate: scan job_posts row: %v", err)
			s.Attempted++
			s.Failed++
			continue
		}
		s.Attempted++

		pbSiteID, ok := siteMap[jp.SiteID]
		if !ok {
			log.Printf("altmigrate: job post id=%d: no PocketBase site for legacy site_id=%d", jp.ID, jp.SiteID)
			s.Failed++
			continue
		}

		existing, err := app.FindRecordsByFilter(
			"jobPosts",
			fmt.Sprintf("jobSiteNumber='%s'&&site='%s'", jp.JobSiteNumber, pbSiteID),
			"", 1, 0,
		)
		if err != nil {
			log.Printf("altmigrate: job post id=%d: existence check: %v", jp.ID, err)
			s.Failed++
			continue
		}
		if len(existing) > 0 {
			s.Skipped++
			continue
		}

		rec := core.NewRecord(col)
		rec.Set("jobSiteNumber", jp.JobSiteNumber)
		rec.Set("site", pbSiteID)
		if t, ok := parsePostedDate(jp.PostedDate); ok {
			rec.Set("postedDate", t)
		}
		rec.Set("content", map[string]any{"title": jp.Title, "body": jp.Body})
		rec.Set("location", map[string]any{"city": jp.City, "country": jp.Country, "suburb": jp.Suburb})

		if err := app.Save(rec); err != nil {
			log.Printf("altmigrate: job post id=%d: save: %v", jp.ID, err)
			s.Failed++
			continue
		}
		s.Written++
	}
	if err := rows.Err(); err != nil {
		return s, fmt.Errorf("iterate job_posts rows: %w", err)
	}

	return s, nil
}

// buildSiteMap returns a mapping of legacy site ID → PocketBase site ID,
// joined by matching site name.
func buildSiteMap(app core.App, db *sql.DB) (map[uint]string, error) {
	rows, err := db.Query("SELECT id, name FROM sites")
	if err != nil {
		return nil, fmt.Errorf("query legacy sites: %w", err)
	}
	defer rows.Close()

	legacyIDToName := map[uint]string{}
	for rows.Next() {
		var id uint
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scan legacy site row: %w", err)
		}
		legacyIDToName[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy sites: %w", err)
	}

	pbSites, err := app.FindRecordsByFilter("sites", "", "", -1, 0)
	if err != nil {
		return nil, fmt.Errorf("load PocketBase sites: %w", err)
	}

	nameToID := make(map[string]string, len(pbSites))
	for _, pbSite := range pbSites {
		nameToID[pbSite.GetString("name")] = pbSite.Id
	}

	siteMap := make(map[uint]string)
	for legacyID, name := range legacyIDToName {
		if pbID, ok := nameToID[name]; ok {
			siteMap[legacyID] = pbID
		}
	}
	return siteMap, nil
}

// parsePostedDate parses a legacy posted_date string. The modernc/sqlite driver
// converts stored TEXT datetimes (e.g. "2024-01-15 14:00:00+00:00") into RFC3339
// format ("2024-01-15T14:00:00Z") when scanning into a string, so both the
// space-separator storage format and RFC3339 scan format are handled.
// Returns false if parsing fails.
func parsePostedDate(s string) (time.Time, bool) {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05-07:00",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	log.Printf("altmigrate: could not parse posted_date %q", s)
	return time.Time{}, false
}
