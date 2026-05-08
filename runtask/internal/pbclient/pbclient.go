package pbclient

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pocketbase "github.com/r--w/pocketbase"
)

// Site represents a job-board site record from PocketBase.
type Site struct {
	Id   string
	Name string
	Url  string
}

// JobPostContent is the JSON object stored in the jobPosts.content field.
type JobPostContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// JobPostLocation is the JSON object stored in the jobPosts.location field.
type JobPostLocation struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Suburb  string `json:"suburb"`
}

// JobPost represents a jobPosts record from PocketBase.
type JobPost struct {
	Id            string
	JobSiteNumber string
	SiteId        string
	Content       JobPostContent
	Location      JobPostLocation
	PostedDate    time.Time
}

// SkillNameWithAliases represents an enabled skillName with its aliases expanded.
type SkillNameWithAliases struct {
	Id      string
	Name    string
	Aliases []string
}

// MonthlyCountReport represents a monthlyCountReports record.
type MonthlyCountReport struct {
	Identifier    string
	YearMonth     string
	YearMonthDate time.Time
	Count         int
	SkillNameId   string
}

// Client is a thin PocketBase API wrapper.
type Client struct {
	pb *pocketbase.Client
}

// New returns a Client configured with user email/password authentication.
func New(url, email, password string) (*Client, error) {
	pb := pocketbase.NewClient(url, pocketbase.WithUserEmailPassword(email, password))
	return &Client{pb: pb}, nil
}

// GetSites returns all records from the sites collection.
func (c *Client) GetSites() ([]Site, error) {
	items, err := getFullList(c.pb, "sites", pocketbase.ParamsList{})
	if err != nil {
		return nil, err
	}
	sites := make([]Site, 0, len(items))
	for _, item := range items {
		sites = append(sites, Site{
			Id:   str(item, "id"),
			Name: str(item, "name"),
			Url:  str(item, "url"),
		})
	}
	return sites, nil
}

// UpsertJobPost creates a jobPost record if one with the same (site, jobSiteNumber) does not
// already exist.
func (c *Client) UpsertJobPost(post JobPost) error {
	existing, err := c.pb.List("jobPosts", pocketbase.ParamsList{
		Filters: fmt.Sprintf(`site = %q && jobSiteNumber = %q`, post.SiteId, post.JobSiteNumber),
		Size:    1,
	})
	if err != nil {
		return fmt.Errorf("list jobPosts: %w", err)
	}
	if existing.TotalItems > 0 {
		return nil // already exists — skip
	}

	contentJSON, _ := json.Marshal(post.Content)
	locationJSON, _ := json.Marshal(post.Location)

	_, err = c.pb.Create("jobPosts", map[string]any{
		"jobSiteNumber": post.JobSiteNumber,
		"site":          post.SiteId,
		"content":       json.RawMessage(contentJSON),
		"location":      json.RawMessage(locationJSON),
		"postedDate":    post.PostedDate,
	})
	return err
}

// GetEnabledSkillNamesWithAliases returns all enabled skillNames with their aliases.
// Uses two calls — one for skill names, one for all aliases — then joins in memory.
func (c *Client) GetEnabledSkillNamesWithAliases() ([]SkillNameWithAliases, error) {
	snItems, err := getFullList(c.pb, "skillNames", pocketbase.ParamsList{
		Filters: `isEnabled = true`,
	})
	if err != nil {
		return nil, fmt.Errorf("list skillNames: %w", err)
	}

	// Build ID→SkillNameWithAliases map.
	byId := make(map[string]*SkillNameWithAliases, len(snItems))
	result := make([]SkillNameWithAliases, 0, len(snItems))
	for _, item := range snItems {
		sn := SkillNameWithAliases{
			Id:   str(item, "id"),
			Name: str(item, "name"),
		}
		result = append(result, sn)
		byId[sn.Id] = &result[len(result)-1]
	}

	// Fetch all aliases and join.
	aliasItems, err := getFullList(c.pb, "skillNameAliases", pocketbase.ParamsList{})
	if err != nil {
		return nil, fmt.Errorf("list skillNameAliases: %w", err)
	}
	for _, item := range aliasItems {
		snId := str(item, "skillName")
		if sn, ok := byId[snId]; ok {
			sn.Aliases = append(sn.Aliases, str(item, "alias"))
		}
	}

	return result, nil
}

// GetAllJobPosts returns all jobPost records with content and location parsed.
func (c *Client) GetAllJobPosts() ([]JobPost, error) {
	items, err := getFullList(c.pb, "jobPosts", pocketbase.ParamsList{})
	if err != nil {
		return nil, err
	}

	posts := make([]JobPost, 0, len(items))
	for _, item := range items {
		jp := JobPost{
			Id:            str(item, "id"),
			JobSiteNumber: str(item, "jobSiteNumber"),
			SiteId:        str(item, "site"),
		}
		// Unmarshal nested JSON content/location.
		if raw, err := json.Marshal(item["content"]); err == nil {
			_ = json.Unmarshal(raw, &jp.Content)
		}
		if raw, err := json.Marshal(item["location"]); err == nil {
			_ = json.Unmarshal(raw, &jp.Location)
		}
		// Parse postedDate. PocketBase returns "2006-01-02 15:04:05.000Z"; normalise
		// the space separator to T so standard RFC3339 parsers handle it.
		if s := str(item, "postedDate"); s != "" {
			s = strings.Replace(s, " ", "T", 1)
			if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
				jp.PostedDate = t
			} else if t, err := time.Parse(time.RFC3339, s); err == nil {
				jp.PostedDate = t
			}
		}
		posts = append(posts, jp)
	}
	return posts, nil
}

// UpsertMonthlyCountReport creates or updates a monthlyCountReports record keyed by identifier.
func (c *Client) UpsertMonthlyCountReport(report MonthlyCountReport) error {
	existing, err := c.pb.List("monthlyCountReports", pocketbase.ParamsList{
		Filters: fmt.Sprintf(`identifier = %q`, report.Identifier),
		Size:    1,
	})
	if err != nil {
		return fmt.Errorf("list monthlyCountReports: %w", err)
	}

	body := map[string]any{
		"identifier":    report.Identifier,
		"YearMonth":     report.YearMonth,
		"yearMonthDate": report.YearMonthDate.Format("2006-01-02 15:04:05.000Z"),
		"count":         report.Count,
		"skillName":     report.SkillNameId,
	}

	if existing.TotalItems > 0 {
		id := str(existing.Items[0], "id")
		return c.pb.Update("monthlyCountReports", id, body)
	}
	_, err = c.pb.Create("monthlyCountReports", body)
	return err
}

// ── helpers ───────────────────────────────────────────────────────────────────

// getFullList paginates through all pages and returns every record.
func getFullList(pb *pocketbase.Client, collection string, params pocketbase.ParamsList) ([]map[string]any, error) {
	const perPage = 200
	params.Size = perPage
	var all []map[string]any

	for page := 1; ; page++ {
		params.Page = page
		resp, err := pb.List(collection, params)
		if err != nil {
			return nil, fmt.Errorf("list %s page %d: %w", collection, page, err)
		}
		all = append(all, resp.Items...)
		if len(all) >= resp.TotalItems {
			break
		}
	}
	return all, nil
}

// str extracts a string field from a record map, returning "" if absent or wrong type.
func str(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}
