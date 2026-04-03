# pocketbaseserver — Change Spec (issue #62)

Extends the base spec at `.ai/base/specs/pocketbaseserver.md`.

## Goals

- Add `roles` and `userRoles` collections so `runtask` and `migrate` can authenticate
  as service accounts with scoped write access.
- Apply PocketBase access rules to all existing collections so only users with the
  correct role can perform writes.

## New collections

### roles

Stores named permission roles. Seeded once during migration; not editable at runtime.

| Field | Type | Constraints |
|---|---|---|
| `name` | text | required, unique |
| `description` | text | required |

Seed records (created in the `up` function of the new migration):

| name | description |
|---|---|
| `webscraper` | Write access to `jobPosts` |
| `reporting` | Write access to `monthlyCountReports` |
| `migration` | Write access to all collections except `users`, `userRoles`, and `roles` |

Access rules:
- List/view: `@request.auth.id != ""`  (any authenticated user may read)
- Create/update/delete: admin only (empty rule — only PocketBase superadmin)

### userRoles

Associates a PocketBase user with one or more roles.

| Field | Type | Constraints |
|---|---|---|
| `user` | relation → users | required |
| `role` | relation → roles | required |

A unique index on `(user, role)` prevents duplicate assignments.

A user with no `userRoles` records is treated as a normal login user.

Access rules:
- List/view: `@request.auth.id != ""`
- Create/update/delete: admin only

## Access rule changes to existing collections

Apply the following write rules. Read rules remain unchanged (public).

| Collection | Write rule |
|---|---|
| `jobPosts` | `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'webscraper' \|\| @collection.userRoles_via_user.role.name ?~ 'migration')` |
| `monthlyCountReports` | `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'reporting' \|\| @collection.userRoles_via_user.role.name ?~ 'migration')` |
| `skillTypes` | `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'migration' \|\| @request.auth.verified = true)` |
| `skillNames` | same as `skillTypes` |
| `skillNameAliases` | same as `skillTypes` |
| `sites` | same as `skillTypes` |

> The `migration` role may write to all of the above. The `webscraper` role may
> write to `jobPosts`; the `reporting` role may write to `monthlyCountReports`.
> Normal authenticated users (verified, no special role) may also write to
> `skillTypes`, `skillNames`, `skillNameAliases`, and `sites`. The frontend views
> for creating and editing skills and skill types have been removed; mutation of
> these collections is performed via the PocketBase admin UI (superadmin, which
> bypasses these rules) or directly through the API by verified users.

## Migration file

Create `pocketbaseserver/migrations/1743552000_add_roles.go`.

- `up`: create `roles` collection, create `userRoles` collection with a unique index
  on `(user, role)`, insert the three seed records into `roles`, apply new access
  rules to existing collections.
- `down`: revert access rules, drop seed records, drop `userRoles`, drop `roles`.

No service account passwords are generated or stored. Operator creates service
accounts manually via the PocketBase admin UI after running the server.

## Integration tests

Write tests in `pocketbaseserver/` before implementing the migration:

- A user with no `userRoles` cannot write to `jobPosts` or `monthlyCountReports`.
- A user with the `webscraper` role can create a `jobPost` record.
- A user with the `reporting` role can create a `monthlyCountReport` record.
- A user with the `migration` role can create records in `sites`, `skillTypes`,
  `skillNames`, `skillNameAliases`, `jobPosts`, and `monthlyCountReports`.
- A user with the `migration` role cannot create or modify `users`, `userRoles`,
  or `roles` records.
- Inserting a duplicate `(user, role)` pair into `userRoles` is rejected.

## Out of scope

- Service account creation (operator responsibility).
- Assigning roles to users (operator responsibility via PocketBase admin UI).
