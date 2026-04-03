package migrator

import (
	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"
)

// Migrator orchestrates the one-shot migration from legacy SQLite to PocketBase.
type Migrator struct {
	db *gorm.DB
	pb *pocketbase.Client
}

// Summary holds per-collection migration results.
type Summary struct {
	Collection string
	Attempted  int
	Written    int
}

// New creates a Migrator.
func New(db *gorm.DB, pb *pocketbase.Client) *Migrator {
	return &Migrator{db: db, pb: pb}
}

// Run executes all migration steps in dependency order and returns per-collection summaries.
func (m *Migrator) Run() ([]Summary, error) {
	var summaries []Summary

	siteIdMap, siteSummary, err := migrateSites(m.db, m.pb)
	summaries = append(summaries, siteSummary)
	if err != nil {
		return summaries, err
	}

	skillTypeIdMap, stSummary, err := migrateSkillTypes(m.db, m.pb)
	summaries = append(summaries, stSummary)
	if err != nil {
		return summaries, err
	}

	skillNameIdMap, snSummary, err := migrateSkillNames(m.db, m.pb, skillTypeIdMap)
	summaries = append(summaries, snSummary)
	if err != nil {
		return summaries, err
	}

	_, snaSummary, err := migrateSkillNameAliases(m.db, m.pb, skillNameIdMap)
	summaries = append(summaries, snaSummary)
	if err != nil {
		return summaries, err
	}

	_, jpSummary, err := migrateJobPosts(m.db, m.pb, siteIdMap)
	summaries = append(summaries, jpSummary)
	if err != nil {
		return summaries, err
	}

	_, mcrSummary, err := migrateMonthlyCountReports(m.db, m.pb, skillNameIdMap)
	summaries = append(summaries, mcrSummary)
	if err != nil {
		return summaries, err
	}

	return summaries, nil
}
