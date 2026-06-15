# Sets Remaining Operations — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `SyncUserSet`, `UpdateUserSetListSet`, and `ReplaceUserSetListSet` to the `rebrickable-go` library and wire them into three new CLI commands in `rebrickable-cli`.

**Architecture:** Changes span two repos. Library work (tasks 1–3) comes first, is tagged as v0.3.0 and pushed to GitHub, then the CLI repo pulls the new version (task 4) and gains the commands (task 5). Each library method follows the existing error-only return pattern for mutating calls; each CLI command follows the existing Cobra pattern in `sets.go`.

**Tech Stack:** Go 1.23+, resty v2 (HTTP), Cobra (CLI), Bazel + Gazelle (CLI build), `httptest` (unit tests).

---

## Repo paths

| Symbol | Path |
|--------|------|
| `$LIB` | `/Users/fabio.mangione/workspace/my-stuff/rebrickable-go` |
| `$CLI` | `/Users/fabio.mangione/workspace/my-stuff/rebrickable-cli` |

---

## Task 1: `SyncUserSet` — library test + implementation

**Files:**
- Modify: `$LIB/client_test.go` (append test)
- Modify: `$LIB/user.go` (append method)

- [ ] **Step 1: Append the failing test to `$LIB/client_test.go`**

Add after `TestDeleteUserSet` (around line 1449):

```go
func TestSyncUserSet(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"synced successfully", 200, false},
		{"server error", 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "token", server.URL)
			err := client.SyncUserSet("10274-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncUserSet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && capturedPath != "/users/token/sets/sync/" {
				t.Errorf("SyncUserSet() path = %q, want /users/token/sets/sync/", capturedPath)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd $LIB && go test ./... -run TestSyncUserSet -v
```

Expected: `FAIL — undefined: Client.SyncUserSet`

- [ ] **Step 3: Append `SyncUserSet` to `$LIB/user.go`**

Add after `DeleteUserSet` (end of file):

```go
func (c *Client) SyncUserSet(setNum string) error {
	resp, err := c.http.R().
		SetBody(map[string]string{"set_num": setNum}).
		Post(c.userPath("/sets/sync/"))
	if err != nil {
		return fmt.Errorf("sync set request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("sync set failed with status %d", resp.StatusCode())
	}
	return nil
}
```

- [ ] **Step 4: Run the test to confirm it passes**

```bash
cd $LIB && go test ./... -run TestSyncUserSet -v
```

Expected: `PASS`

- [ ] **Step 5: Run the full test suite**

```bash
cd $LIB && go test ./...
```

Expected: all tests pass, no failures.

- [ ] **Step 6: Commit**

```bash
cd $LIB
git add user.go client_test.go
git commit -m "feat: add SyncUserSet — POST /users/{token}/sets/sync/"
```

---

## Task 2: `UpdateUserSetListSet` + `ReplaceUserSetListSet` — library test + implementation

**Files:**
- Modify: `$LIB/client_test.go` (append two tests)
- Modify: `$LIB/user.go` (append two methods)

- [ ] **Step 1: Append the failing tests to `$LIB/client_test.go`**

Add after `TestSyncUserSet`:

```go
func TestUpdateUserSetListSet(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"updated successfully", 200, false},
		{"not found", 404, true},
		{"server error", 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "token", server.URL)
			err := client.UpdateUserSetListSet("42", "10274-1", 2, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserSetListSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReplaceUserSetListSet(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"replaced successfully", 200, false},
		{"not found", 404, true},
		{"server error", 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "token", server.URL)
			err := client.ReplaceUserSetListSet("42", "10274-1", 2, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceUserSetListSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Run the tests to confirm they fail**

```bash
cd $LIB && go test ./... -run "TestUpdateUserSetListSet|TestReplaceUserSetListSet" -v
```

Expected: `FAIL — undefined: Client.UpdateUserSetListSet`

- [ ] **Step 3: Append both methods to `$LIB/user.go`**

Add after `SyncUserSet`:

```go
func (c *Client) UpdateUserSetListSet(listID, setNum string, quantity int, includeSpares bool) error {
	resp, err := c.http.R().
		SetBody(map[string]interface{}{"quantity": quantity, "include_spares": includeSpares}).
		Patch(c.userPath(fmt.Sprintf("/setlists/%s/sets/%s/", listID, setNum)))
	if err != nil {
		return fmt.Errorf("update set list set request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("update set list set failed with status %d", resp.StatusCode())
	}
	return nil
}

func (c *Client) ReplaceUserSetListSet(listID, setNum string, quantity int, includeSpares bool) error {
	resp, err := c.http.R().
		SetBody(map[string]interface{}{"quantity": quantity, "include_spares": includeSpares}).
		Put(c.userPath(fmt.Sprintf("/setlists/%s/sets/%s/", listID, setNum)))
	if err != nil {
		return fmt.Errorf("replace set list set request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("replace set list set failed with status %d", resp.StatusCode())
	}
	return nil
}
```

- [ ] **Step 4: Run the tests to confirm they pass**

```bash
cd $LIB && go test ./... -run "TestUpdateUserSetListSet|TestReplaceUserSetListSet" -v
```

Expected: `PASS`

- [ ] **Step 5: Run the full test suite**

```bash
cd $LIB && go test ./...
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd $LIB
git add user.go client_test.go
git commit -m "feat: add UpdateUserSetListSet (PATCH) and ReplaceUserSetListSet (PUT)"
```

---

## Task 3: Tag and push `v0.3.0` in rebrickable-go

**Files:** none (git operations only)

- [ ] **Step 1: Tag the release**

```bash
cd $LIB && git tag v0.3.0
```

- [ ] **Step 2: Push main and tags**

```bash
cd $LIB && git push origin main --tags
```

- [ ] **Step 3: Verify the tag appears on GitHub**

```bash
gh release list --repo bu3/rebrickable-go | head -5
```

Expected: `v0.3.0` in the list (or check `git ls-remote --tags origin`).

---

## Task 4: Update `rebrickable-cli` to use `rebrickable-go v0.3.0`

**Files:**
- Modify: `$CLI/go.mod`
- Modify: `$CLI/go.sum`
- Modify: `$CLI/deps.bzl`

- [ ] **Step 1: Upgrade the dependency**

```bash
cd $CLI && go get github.com/bu3/rebrickable-go@v0.3.0 && go mod tidy
```

Expected: `go: upgraded github.com/bu3/rebrickable-go v0.2.0 => v0.3.0`

- [ ] **Step 2: Regenerate Bazel repo mappings**

```bash
cd $CLI && bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies
```

Expected: exits 0, `deps.bzl` now shows `version = "v0.3.0"` for `com_github_bu3_rebrickable_go`.

- [ ] **Step 3: Verify the build still compiles**

```bash
cd $CLI && bazel build //cli
```

Expected: `Build completed successfully`.

- [ ] **Step 4: Commit**

```bash
cd $CLI
git add go.mod go.sum deps.bzl
git commit -m "chore: bump rebrickable-go to v0.3.0"
```

---

## Task 5: Add CLI commands — `sync`, `update`, `replace`

**Files:**
- Modify: `$CLI/cli/cmd/sets.go`

- [ ] **Step 1: Add the `includeSpares` package-level variable**

In `$CLI/cli/cmd/sets.go`, add `includeSpares` to the existing var block at the top (lines 12–15):

```go
var setNumber string
var setListName string
var setListID string
var quantity int
var includeSpares bool
```

- [ ] **Step 2: Register `syncSetsCmd` in `setCommands()`**

The current `setCommands()` function (around line 23) ends with:

```go
	saveSetsCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	deleteSetsCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	getSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	replaceSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	replaceSetCmd.Flags().IntVarP(&quantity, "quantity", "q", 1, "Quantity")
```

Replace that block with:

```go
	setsCmd.AddCommand(syncSetsCmd)

	saveSetsCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	deleteSetsCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	getSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	replaceSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	replaceSetCmd.Flags().IntVarP(&quantity, "quantity", "q", 1, "Quantity")
	syncSetsCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
```

- [ ] **Step 3: Register `updateSetListSetCmd` and `replaceSetListSetCmd` in `setListSetsCommands()`**

The current `setListSetsCommands()` function (around line 56) ends with:

```go
	deleteSetListSetCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	deleteSetListSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
```

Replace that block with:

```go
	setListSetsCmd.AddCommand(updateSetListSetCmd)
	setListSetsCmd.AddCommand(replaceSetListSetCmd)

	getSetListSetsCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	getSetListSetCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	getSetListSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	saveSetListSetCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	saveSetListSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	deleteSetListSetCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	deleteSetListSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	updateSetListSetCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	updateSetListSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	updateSetListSetCmd.Flags().IntVarP(&quantity, "quantity", "q", 1, "Quantity")
	updateSetListSetCmd.Flags().BoolVar(&includeSpares, "include_spares", false, "Include spare parts")
	replaceSetListSetCmd.Flags().StringVarP(&setListID, "set_list_id", "l", "", "Set List id")
	replaceSetListSetCmd.Flags().StringVarP(&setNumber, "set_num", "n", "", "Set number")
	replaceSetListSetCmd.Flags().IntVarP(&quantity, "quantity", "q", 1, "Quantity")
	replaceSetListSetCmd.Flags().BoolVar(&includeSpares, "include_spares", false, "Include spare parts")
```

- [ ] **Step 4: Append the three new command vars to the bottom of `sets.go`**

Add after `adjustedSetNumber()` (end of file):

```go
var syncSetsCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient(cmd)
		if err := client.SyncUserSet(adjustedSetNumber()); err != nil {
			return err
		}
		fmt.Printf("Synced set: %s\n", adjustedSetNumber())
		return nil
	},
}

var updateSetListSetCmd = &cobra.Command{
	Use:   "update",
	Short: "update",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient(cmd)
		if err := client.UpdateUserSetListSet(setListID, adjustedSetNumber(), quantity, includeSpares); err != nil {
			return err
		}
		fmt.Println("Updated set in set list")
		return nil
	},
}

var replaceSetListSetCmd = &cobra.Command{
	Use:   "replace",
	Short: "replace",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient(cmd)
		if err := client.ReplaceUserSetListSet(setListID, adjustedSetNumber(), quantity, includeSpares); err != nil {
			return err
		}
		fmt.Println("Replaced set in set list")
		return nil
	},
}
```

- [ ] **Step 5: Build to verify no compile errors**

```bash
cd $CLI && bazel build //cli
```

Expected: `Build completed successfully`.

- [ ] **Step 6: Smoke-test the new commands appear in help**

```bash
cd $CLI && .bazel/bin/cli/cli_/cli user sets --help
```

Expected: `sync` listed as a subcommand.

```bash
cd $CLI && .bazel/bin/cli/cli_/cli user setListSets --help
```

Expected: `update` and `replace` listed as subcommands.

- [ ] **Step 7: Commit**

```bash
cd $CLI
git add cli/cmd/sets.go
git commit -m "feat: add user sets sync, setListSets update/replace commands"
```

---

## Task 6: Run full test suite

- [ ] **Step 1: Run all tests**

```bash
cd $CLI && ./test.sh
```

Expected: all tests pass. Unit tests (rebrickable-go) and integration tests (txtar via Bazel) both green.

- [ ] **Step 2: If integration tests fail with 429**

The retry logic (added in v0.2.0) will wait up to 10s and retry 3 times automatically. If still failing, wait 60 seconds and re-run `./test.sh`. This is a transient rate-limit issue, not a code bug.
