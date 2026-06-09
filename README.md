# rebrickable-go

Go client library for the [Rebrickable API](https://rebrickable.com/api/).

## Installation

```bash
go get github.com/bu3/rebrickable-go
```

## Usage

### LEGO catalog (no authentication required)

```go
client := rebrickable.NewClient("your-api-key")

sets, err := client.GetLegoSets()
set, err := client.GetLegoSet("10497-1")
colors, err := client.GetLegoColors()
themes, err := client.GetLegoThemes()
```

### User collection (authentication required)

```go
client, err := rebrickable.NewAuthenticatedClient("your-api-key", "username", "password")
if err != nil {
    log.Fatal(err)
}

sets, err := client.GetUserSets()
err = client.StoreUserSet("10497-1")
err = client.DeleteUserSet("10497-1")
```

## API

### Constructors

| Function | Description |
|----------|-------------|
| `NewClient(apiKey string) *Client` | Anonymous client for read-only LEGO catalog endpoints |
| `NewAuthenticatedClient(apiKey, username, password string) (*Client, error)` | Authenticated client for user collection endpoints |

### LEGO catalog (`lego.go`)

| Method | Description |
|--------|-------------|
| `GetLegoSets()` | List all sets |
| `GetLegoSet(setNum)` | Get a single set |
| `GetLegoSetAlternates(setNum)` | List alternate builds for a set |
| `GetLegoSetMinifigs(setNum)` | List minifigs in a set |
| `GetLegoSetParts(setNum)` | List parts in a set |
| `GetLegoSetSets(setNum)` | List sub-sets within a set |
| `GetLegoColors()` | List all colors |
| `GetLegoColor(id)` | Get a single color |
| `GetLegoElement(elementID)` | Get a single element |
| `GetLegoMinifigs()` | List all minifigs |
| `GetLegoMinifig(figNum)` | Get a single minifig |
| `GetLegoMinifigParts(figNum)` | List parts of a minifig |
| `GetLegoMinifigSets(figNum)` | List sets containing a minifig |
| `GetLegoPartCategories()` | List all part categories |
| `GetLegoPartCategory(id)` | Get a single part category |
| `GetLegoThemes()` | List all themes |
| `GetLegoTheme(id)` | Get a single theme |

### LEGO parts (`lego_parts.go`)

| Method | Description |
|--------|-------------|
| `GetLegoParts(filter PartsFilter)` | List parts with optional filters |
| `GetLegoPart(partNum)` | Get a single part |
| `GetLegoPartColors(partNum)` | List colors a part has appeared in |
| `GetLegoPartColor(partNum, colorID)` | Get a specific part/color combination |
| `GetLegoPartColorSets(partNum, colorID)` | List sets containing a part/color combination |

`PartsFilter` fields: `PartNum`, `PartNums`, `PartCatID`, `ColorID`, `BricklinkID`, `BrickowlID`, `LegoID`, `LdrawID`, `Ordering`, `Search` (all strings, all optional).

### User collection (`user.go`)

| Method | Description |
|--------|-------------|
| `GetUserSets()` | List sets in the user's collection |
| `GetUserSet(setNum)` | Get a single set from the collection |
| `StoreUserSet(setNum)` | Add a set to the collection |
| `ReplaceUserSet(setNum, quantity)` | Update a set's quantity |
| `DeleteUserSet(setNum)` | Remove a set (404 → nil, idempotent) |
| `GetUserSetLists()` | List all set lists |
| `GetUserSetList(listID)` | Get a single set list |
| `StoreUserSetList(name)` | Create a set list |
| `UpdateUserSetList(listID, name)` | Partially update a set list |
| `ReplaceUserSetList(listID, name)` | Replace a set list |
| `DeleteUserSetList(id)` | Delete a set list (404 → nil, idempotent) |
| `GetUserSetListSets(listID)` | List sets within a set list |
| `GetUserSetListSet(listID, setNum)` | Get a single set from a set list |
| `StoreUserSetListSet(listID, setNum)` | Add a set to a set list |
| `DeleteUserSetListSet(listID, setNum)` | Remove a set from a set list (404 → nil, idempotent) |

## Requirements

- Go 1.23+
- A Rebrickable API key — get one at [rebrickable.com/api](https://rebrickable.com/api/)

## Testing

```bash
go test ./...
```
