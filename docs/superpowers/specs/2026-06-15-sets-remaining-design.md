# Sets Remaining Operations — Design

**Goal:** Implement the three remaining sets-related endpoints across `rebrickable-go` (library) and `rebrickable-cli` (CLI).

**Endpoints covered:**
- `POST /api/v3/users/{user_token}/sets/sync/`
- `PATCH /api/v3/users/{user_token}/setlists/{list_id}/sets/{set_num}/`
- `PUT /api/v3/users/{user_token}/setlists/{list_id}/sets/{set_num}/`

---

## Library (`rebrickable-go/user.go`)

Three new methods, all following the existing `error`-only return pattern for mutating calls.

### `SyncUserSet`

```go
func (c *Client) SyncUserSet(setNum string) error
```

- `POST /users/{token}/sets/sync/`
- Body: `set_num` (form data)
- Success: HTTP 200 (sync is idempotent — upserts the set regardless of whether it already exists)
- Error: wraps non-200 status in a descriptive error

Mirrors `StoreUserSet` but targets `/sets/sync/` and expects 200 instead of 201.

### `UpdateUserSetListSet`

```go
func (c *Client) UpdateUserSetListSet(listID, setNum string, quantity int, includeSpares bool) error
```

- `PATCH /users/{token}/setlists/{list_id}/sets/{set_num}/`
- Body: `quantity`, `include_spares` (form data, both optional per spec)
- Success: HTTP 200

### `ReplaceUserSetListSet`

```go
func (c *Client) ReplaceUserSetListSet(listID, setNum string, quantity int, includeSpares bool) error
```

- `PUT /users/{token}/setlists/{list_id}/sets/{set_num}/`
- Body: `quantity`, `include_spares` (form data)
- Success: HTTP 200

---

## CLI (`rebrickable-cli/cli/cmd/sets.go`)

### New commands

| Command | Flags | Calls |
|---------|-------|-------|
| `user sets sync` | `-n <set_num>` | `SyncUserSet` |
| `user setListSets update` | `-l <list_id>` `-n <set_num>` `-q <quantity>` `--include_spares` | `UpdateUserSetListSet` |
| `user setListSets replace` | `-l <list_id>` `-n <set_num>` `-q <quantity>` `--include_spares` | `ReplaceUserSetListSet` |

### Flag details

- `-n` / `--set_num`: reuses existing `setNumber` package-level var and `adjustedSetNumber()` helper (appends `-1` suffix if missing)
- `-l` / `--set_list_id`: reuses existing `setListID` var
- `-q` / `--quantity`: reuses existing `quantity` var (default `1`)
- `--include_spares`: new `bool` flag, package-level var `includeSpares`, scoped to `update` and `replace` only

### Output strings

- `user sets sync` → `"Synced set: <set_num>"`
- `user setListSets update` → `"Updated set in set list"`
- `user setListSets replace` → `"Replaced set in set list"`

### Registration

`syncSetsCmd` is added to `setsCmd` inside `setCommands()`.
`updateSetListSetCmd` and `replaceSetListSetCmd` are added to `setListSetsCmd` inside `setListSetsCommands()`.

---

## Testing

### Library unit tests (`rebrickable-go/client_test.go`)

Table-driven tests using `httptest.NewServer` + `newClientWithBaseURL` for each new method:

- Verify correct HTTP method and path
- Verify request body contains expected fields
- Verify nil returned on expected success status
- Verify error returned on unexpected status

### CLI integration tests

No new txtar files:
- `sync` mutates the default set list — the existing `sets.txtar` already covers the sets group adequately; adding sync would require careful cleanup
- `UpdateUserSetListSet` / `ReplaceUserSetListSet` operate on setlists with auto-generated IDs, which per project convention are not suitable for txtar tests

---

## Implementation notes

- `SyncUserSet` sends `set_num` as form data (same as `StoreUserSet`); expects HTTP 200, not 201
- `UpdateUserSetListSet` and `ReplaceUserSetListSet` send `quantity` and `include_spares` as form data using `map[string]interface{}`
- Both setListSets methods follow the `ReplaceUserSet` pattern for body construction
- No new types needed — no response body is returned by any of the three endpoints
