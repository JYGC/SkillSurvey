# frontend — Change Spec (issue #62)

Extends the base spec at `.ai/base/specs/frontend.md`.

## Goals

- Remove all `fetch()` calls to `:3000`; replace with PocketBase JS SDK calls to
  `:8090`.
- Update TypeScript interfaces in `skills.ts` to match PocketBase's string IDs and
  camelCase field names.
- Fetch monthly count report data from the `monthlyCountReports` PocketBase
  collection instead of the legacy `/report/getmonthlycount` endpoint.

## Interface changes (`src/schemas/skills.ts`)

Replace the existing numeric-ID GORM-shaped interfaces with PocketBase-shaped ones:

```typescript
export interface SkillType {
  id: string;
  name: string;
  description: string;
  expand?: { skillNames_via_skillType?: SkillName[] };
}

export interface SkillName {
  id: string;
  skillType: string;          // relation ID
  name: string;
  isEnabled: boolean;
  expand?: {
    skillType?: SkillType;
    skillNameAliases_via_skillName?: SkillNameAlias[];
  };
}

export interface SkillNameAlias {
  id: string;
  skillName: string;          // relation ID
  alias: string;
  expand?: { skillName?: SkillName };
}
```

## Files to delete

The following views and components are removed entirely. All skill and skill-type
management (list, create, edit, delete) is handled via the PocketBase admin UI instead.

- `src/views/public/SkillTypeAdd.vue`
- `src/views/public/SkillTypeEdit.vue`
- `src/views/public/SkillAdd.vue`
- `src/views/public/SkillEdit.vue`
- `src/views/public/SkillList.vue`
- `src/views/public/SkillTypeList.vue`
- `src/components/SkillView.vue` (only used by `SkillAdd.vue` and `SkillEdit.vue`)
- `src/components/SkillTypeView.vue` (only used by `SkillTypeAdd.vue` and `SkillTypeEdit.vue`)

Remove any Vue Router routes and navigation links that reference these deleted views.

## View-by-view migration

### `MonthlyCountReport.vue`

- Remove `fetch('http://localhost:3000/report/getmonthlycount')`.
- Replace with:
  ```typescript
  getBackendClient().collection('monthlyCountReports').getFullList({
    expand: 'skillName',
    sort: 'yearMonthDate',
  })
  ```
- The chart dataset construction must be adapted to the PocketBase record shape.
  The `monthlyCountReports` collection has two time-related fields: `YearMonth`
  (text, e.g. `"2024-01"`) used as the human-readable label and grouping key, and
  `yearMonthDate` (date, first day of the period) used for chronological sorting.
  Access them as `record.YearMonth` (string), `record.count` (number), and
  `record.expand?.skillName?.name`.
- Filter to the most recent 12 distinct `YearMonth` values in the component before
  passing data to Chart.js.

## Non-functional requirements

- All hardcoded `http://localhost:3000` strings must be removed; no replacement
  environment variable is needed because all data now comes from `VUE_APP_POCKETBASE_URL`.
- `npm run lint` must pass with no errors after the changes.
- No new npm dependencies.

## Out of scope

- Authentication flow changes (Login, RegisterUser, Settings, Profile are unchanged).
- Server-side rendering.
