package report

import (
	"fmt"
	"strings"
	"time"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/pbclient"
)

// Run computes monthly skill-demand counts from jobPosts and upserts monthlyCountReports.
// Only jobPosts from the 13-month window ending now (1 year + 1 month) are considered,
// keeping the query and in-memory footprint bounded as the collection grows.
func Run(_ config.Config, pb *pbclient.Client) error {
	skillNames, err := pb.GetEnabledSkillNamesWithAliases()
	if err != nil {
		return fmt.Errorf("get skill names: %w", err)
	}

	since := time.Now().AddDate(-1, -1, 0)
	jobPosts, err := pb.GetAllJobPosts(since)
	if err != nil {
		return fmt.Errorf("get job posts: %w", err)
	}

	// Build a map of yearMonth → jobPosts for efficient grouping.
	postsByMonth := make(map[string][]pbclient.JobPost)
	for _, jp := range jobPosts {
		ym := jp.PostedDate.Format("2006-01")
		if ym == "0001-01" {
			// PostedDate was not parsed — skip.
			continue
		}
		postsByMonth[ym] = append(postsByMonth[ym], jp)
	}

	for _, sn := range skillNames {
		for yearMonth, posts := range postsByMonth {
			count := 0
			for _, jp := range posts {
				body := strings.ToLower(jp.Content.Body)
				if bodyMatchesAnyAlias(body, sn.Name, sn.Aliases) {
					count++
				}
			}
			if count == 0 {
				continue
			}

			ymDate, err := time.Parse("2006-01", yearMonth)
			if err != nil {
				continue
			}

			identifier := fmt.Sprintf("%s_%s", sn.Id, yearMonth)
			if err := pb.UpsertMonthlyCountReport(pbclient.MonthlyCountReport{
				Identifier:    identifier,
				YearMonth:     yearMonth,
				YearMonthDate: ymDate,
				Count:         count,
				SkillNameId:   sn.Id,
			}); err != nil {
				fmt.Printf("upsert monthlyCountReport %s: %v\n", identifier, err)
			}
		}
	}
	return nil
}

// bodyMatchesAnyAlias reports whether body contains the skill name or any of its aliases
// using the 16 word-boundary patterns ported from backend/internal/database/jobposttablecall.go.
func bodyMatchesAnyAlias(body, skillName string, aliases []string) bool {
	terms := append([]string{skillName}, aliases...)
	for _, term := range terms {
		if matchesTerm(body, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

// matchesTerm checks the 16 word-boundary combinations for term within body.
// Patterns: { , . \n} × term × { , . \n}
func matchesTerm(body, term string) bool {
	separators := []string{" ", ",", ".", "\n"}
	for _, prefix := range separators {
		for _, suffix := range separators {
			if strings.Contains(body, prefix+term+suffix) {
				return true
			}
		}
	}
	return false
}
