package migrations

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return seedSkillData(app)
	}, func(app core.App) error {
		return removeSeededSkillData(app)
	})
}

// oldID values are the numeric primary keys from the legacy SQLite database.
// They are used only within this file to wire up relations during seeding.

var seedSkillTypes = []struct {
	oldID       int
	name        string
	description string
}{
	{1, "Back End Language", ""},
	{2, "Back End Framework", ""},
	{3, "Front End Language", ""},
	{4, "Front End Framework", ""},
	{5, "Database", ""},
	{6, "Hosting Platform", ""},
	{7, "Business Platform", ""},
	{8, "Source Control", ""},
	{9, "Concepts", ""},
	{10, "Testing Framework", "Includes unit testing"},
	{11, "Stylesheet Language", ""},
	{12, "Mobile Framework", ""},
	{13, "Front End Tools", ""},
	{14, "API Calls", "Calls to API"},
	{15, "Cloud Monitoring", ""},
	{16, "Operating System", ""},
	{17, "Development Framework", ""},
	{18, "Development Practices", "example: Test Driving Development, Continuous Improvement/Continuous Development"},
	{19, "Object-Relational Mapping", "Also known as ORMs. Query and manipulate data from a database using an object-oriented paradigm"},
	{20, "Artificial Intelligence", "Artificial Intelligence"},
	{21, "BaaS", "Backend As A Service"},
	{22, "AI Editors", "Cursor, ChatGPT for Code"},
}

var seedSkillNames = []struct {
	oldID          int
	skillTypeOldID int
	name           string
	isEnabled      bool
}{
	{1, 2, ".NET Framework", true},
	{2, 2, "ASP.NET MVC", true},
	{3, 2, "ASP.NET WebAPI 2", true},
	{4, 2, "ASP.NET WebForms", true},
	{5, 6, "Amazon Web Services", true},
	{6, 4, "Angular", true},
	{7, 6, "Azure", true},
	{8, 1, "C#", true},
	{9, 1, "C++", true},
	{10, 19, "Entity Framework", true},
	{11, 8, "Git", true},
	{12, 1, "Java", true},
	{13, 3, "JavaScript", true},
	{14, 1, "Microsoft SQL", true},
	{15, 1, "MySQL", true},
	{16, 1, "Node", true},
	{17, 7, "Power BI", true},
	{18, 4, "React", true},
	{19, 1, "SQLite", true},
	{20, 8, "Team Foundation Server", true},
	{21, 9, "Internet of Things", true},
	{22, 2, ".NET Core", true},
	{23, 1, "Python", true},
	{24, 1, "PHP", true},
	{25, 10, "NUnit", true},
	{26, 10, "JUnit", true},
	{27, 10, "PHPUnit", true},
	{28, 1, "Go", true},
	{29, 2, "Flask", true},
	{30, 2, "Django", true},
	{31, 4, "Vue.js", true},
	{32, 11, "CSS", true},
	{33, 11, "SASS", true},
	{34, 11, "LESS CSS", true},
	{35, 11, "Stylus", true},
	{36, 3, "Typescript", true},
	{37, 12, "Xamarin", true},
	{38, 5, "PostgreSQL", true},
	{39, 1, "Ruby", true},
	{40, 1, "Ruby On Rails", true},
	{41, 1, "Kotlin", true},
	{42, 4, "Bootstrap", true},
	{43, 13, "Webpack", true},
	{44, 14, "GraphQL", true},
	{45, 5, "Redis", true},
	{46, 6, "Docker", true},
	{47, 2, "FastAPI", true},
	{48, 5, "MongoDB", true},
	{49, 15, "DataDog", true},
	{50, 15, "CloudWatch", true},
	{51, 10, "Jest", true},
	{52, 16, "Graphene", true},
	{53, 1, "Elixir", true},
	{54, 17, "Next.js", true},
	{55, 17, "Nuxt", true},
	{56, 17, "SvelteKit", true},
	{57, 17, "Angular Universal", true},
	{58, 4, "Svelte", true},
	{59, 1, "Perl", true},
	{60, 18, "Test Driven Development", true},
	{61, 18, "Continuous integration and continuous delivery", true},
	{62, 18, "Behavior Driven Development", true},
	{63, 11, "Flexbox", true},
	{64, 19, "NHibernate", true},
	{65, 10, "Selenium", true},
	{66, 4, "Blazor", true},
	{67, 11, "Tailwind CSS", true},
	{68, 20, "GPT", true},
	{69, 20, "ChatGPT", true},
	{70, 1, "F#", true},
	{71, 1, "OCaml", true},
	{72, 14, "REST", true},
	{73, 14, "SOAP", true},
	{74, 14, "gRPC", true},
	{75, 14, "Websockets", true},
	{76, 20, "CoPilot", true},
	{77, 21, "Convex", true},
	{78, 21, "Firebase", true},
	{79, 21, "Supabase", true},
	{80, 21, "AWS Amplify", true},
	{81, 21, "Appwrite", true},
	{82, 21, "Nhost", true},
	{83, 21, "Backendless", true},
	{84, 21, "Back4App", true},
	{85, 20, "Grok", true},
	{86, 20, "Claude", true},
	{87, 20, "DeepSeek", true},
	{88, 22, "Cursor", true},
	{89, 22, "Zed", true},
	{90, 22, "Aider", true},
	{91, 22, "CodeEdit", true},
	{92, 22, "Roo Code", true},
	{93, 22, "Kiro", true},
	{94, 22, "Windsurf", true},
}

// skillNameOldID=0 in the source data is an invalid reference — those aliases are omitted.
var seedSkillNameAliases = []struct {
	skillNameOldID int
	alias          string
}{
	{1, ".NET"},
	{1, "NET"},
	{2, "MVC"},
	{3, "ASP.NET WebAPI"},
	{3, "Web API 2"},
	{3, "ASP.NET Web API"},
	{3, "WebAPI"},
	{3, "Web API"},
	{4, "WebForms"},
	{4, "Web Forms"},
	{5, "AWS"},
	{6, "Angular Js"},
	{6, "AngularJs"},
	{8, "C Sharp"},
	{8, "CS"},
	{10, "EF"},
	{10, "EntityFramework"},
	{11, "Git Hub"},
	{11, "GitHub"},
	{13, "JS"},
	{13, "Java Script"},
	{14, "MS SQL"},
	{14, "MSSQL"},
	{14, "T-SQL"},
	{15, "My SQL"},
	{16, "Node js"},
	{16, "Node.js"},
	{16, "NodeJS"},
	{17, "BI"},
	{17, "PowerBI"},
	{18, "React Js"},
	{18, "ReactJs"},
	{20, "TFS"},
	{21, "IoT"},
	{22, "NETCore"},
	{22, ".NETCore"},
	{22, "CoreCLR"},
	{22, "NET Core"},
	{22, "Core .NET"},
	{22, "Core CLR"},
	{31, "Vue"},
	{31, "Vue Js"},
	{31, "VueJs"},
	{32, "C S S"},
	{33, "S A S S"},
	{33, "SCSS"},
	{34, "LESS and SASS"},
	{34, "SASS and LESS"},
	{34, "LESS/SASS"},
	{34, "Sass or LESS"},
	{34, "SASS/LESS/Stylus"},
	{34, "LESS or Sass"},
	{34, "less, sass"},
	{34, "sass, less"},
	{34, "Stylus/LESS"},
	{36, "Type Script"},
	{38, "Postgres"},
	{38, "Postgres Sql"},
	{38, "PostgresSql"},
	{38, "Postgres Server"},
	{38, "Postgres Sql Server"},
	{38, "Pgsql"},
	{38, "pg sql"},
	{40, "Rails"},
	{40, "RoR"},
	{43, "Web pack"},
	{43, "Web-pack"},
	{44, "Graph QL"},
	{44, "Graph-QL"},
	{47, "Fast API"},
	{48, "Mongo"},
	{49, "Data Dog"},
	{49, "Data-Dog"},
	{50, "Cloud Watch"},
	{54, "Next Js"},
	{56, "Svelte Kit"},
	{60, "TDD"},
	{61, "CI/CD"},
	{61, "CI"},
	{61, "Continuous Integration/Continuous Delivery"},
	{61, "Continuous integration"},
	{61, "Continuous delivery"},
	{62, "BDD"},
	{62, "Behaviour Driven Development"},
	{63, "Flex box"},
	{65, "Selenium WebDriver"},
	{66, ".NET Blazor"},
	{67, "Tailwind"},
	{68, "ChatGPT"},
	{69, "Chat GPT"},
	{70, "F sharp"},
	{70, "FSharp"},
	{70, "F #"},
	{71, "O Caml"},
	{74, "G RPC"},
	{76, "Co Pilot"},
	{76, "Co-Pilot"},
	{78, "Fire base"},
	{79, "Supa base"},
	{80, "Amplify"},
	{81, "App write"},
	{82, "N Host"},
	{85, "X AI"},
	{85, "GrokAI"},
	{87, "Deep Seek"},
	{92, "RooCode"},
	{94, "Wind surf"},
}

func seedSkillData(app core.App) error {
	// description was marked Required in init_collections but most skill types have none —
	// relax the constraint before seeding so empty strings are accepted.
	if err := setSkillTypeDescriptionRequired(app, false); err != nil {
		return err
	}

	skillTypeIDMap := make(map[int]string, len(seedSkillTypes))

	skillTypesCol, err := app.FindCollectionByNameOrId("skillTypes")
	if err != nil {
		return fmt.Errorf("find skillTypes collection: %w", err)
	}
	for _, st := range seedSkillTypes {
		if existing, err := app.FindFirstRecordByData("skillTypes", "name", st.name); err == nil {
			skillTypeIDMap[st.oldID] = existing.Id
			continue
		}
		rec := core.NewRecord(skillTypesCol)
		rec.Set("name", st.name)
		rec.Set("description", st.description)
		if err := app.Save(rec); err != nil {
			return fmt.Errorf("save skillType %q: %w", st.name, err)
		}
		skillTypeIDMap[st.oldID] = rec.Id
	}

	skillNameIDMap := make(map[int]string, len(seedSkillNames))

	skillNamesCol, err := app.FindCollectionByNameOrId("skillNames")
	if err != nil {
		return fmt.Errorf("find skillNames collection: %w", err)
	}
	for _, sn := range seedSkillNames {
		skillTypeID, ok := skillTypeIDMap[sn.skillTypeOldID]
		if !ok {
			return fmt.Errorf("skillType oldID %d not resolved for skillName %q", sn.skillTypeOldID, sn.name)
		}
		if existing, err := app.FindFirstRecordByData("skillNames", "name", sn.name); err == nil {
			skillNameIDMap[sn.oldID] = existing.Id
			continue
		}
		rec := core.NewRecord(skillNamesCol)
		rec.Set("name", sn.name)
		rec.Set("isEnabled", sn.isEnabled)
		rec.Set("skillType", skillTypeID)
		if err := app.Save(rec); err != nil {
			return fmt.Errorf("save skillName %q: %w", sn.name, err)
		}
		skillNameIDMap[sn.oldID] = rec.Id
	}

	skillNameAliasesCol, err := app.FindCollectionByNameOrId("skillNameAliases")
	if err != nil {
		return fmt.Errorf("find skillNameAliases collection: %w", err)
	}
	for _, a := range seedSkillNameAliases {
		skillNameID, ok := skillNameIDMap[a.skillNameOldID]
		if !ok {
			return fmt.Errorf("skillName oldID %d not resolved for alias %q", a.skillNameOldID, a.alias)
		}
		existing, _ := app.FindFirstRecordByFilter(
			"skillNameAliases",
			"skillName = {:sn} && alias = {:al}",
			dbx.Params{"sn": skillNameID, "al": a.alias},
		)
		if existing != nil {
			continue
		}
		rec := core.NewRecord(skillNameAliasesCol)
		rec.Set("skillName", skillNameID)
		rec.Set("alias", a.alias)
		if err := app.Save(rec); err != nil {
			return fmt.Errorf("save alias %q for skillName oldID %d: %w", a.alias, a.skillNameOldID, err)
		}
	}

	return nil
}

func setSkillTypeDescriptionRequired(app core.App, required bool) error {
	col, err := app.FindCollectionByNameOrId("skillTypes")
	if err != nil {
		return fmt.Errorf("find skillTypes collection: %w", err)
	}
	f := col.Fields.GetByName("description")
	if f == nil {
		return fmt.Errorf("skillTypes.description field not found")
	}
	tf, ok := f.(*core.TextField)
	if !ok {
		return fmt.Errorf("skillTypes.description is not a TextField")
	}
	tf.Required = required
	return app.Save(col)
}

func removeSeededSkillData(app core.App) error {
	for _, a := range seedSkillNameAliases {
		recs, err := app.FindAllRecords("skillNameAliases", dbx.HashExp{"alias": a.alias})
		if err != nil {
			continue
		}
		for _, rec := range recs {
			_ = app.Delete(rec)
		}
	}

	for _, sn := range seedSkillNames {
		rec, err := app.FindFirstRecordByData("skillNames", "name", sn.name)
		if err != nil {
			continue
		}
		_ = app.Delete(rec)
	}

	for _, st := range seedSkillTypes {
		rec, err := app.FindFirstRecordByData("skillTypes", "name", st.name)
		if err != nil {
			continue
		}
		_ = app.Delete(rec)
	}

	return nil
}
