package rebrickable

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_SetsAuthorizationHeader(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newClientWithBaseURL("mykey", "", server.URL)
	_, _, _ = fetchAllPages[struct{}](c.http, "/")
	if capturedAuth != "key mykey" {
		t.Errorf("Authorization = %q, want %q", capturedAuth, "key mykey")
	}
}

func TestGetUserToken_Success(t *testing.T) {
	type tokenResp struct {
		UserToken string `json:"user_token"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(tokenResp{UserToken: "tok123"})
	}))
	defer server.Close()

	c := newClientWithBaseURL("apikey", "", server.URL)
	token, err := c.getUserToken("user", "pass")
	if err != nil {
		t.Fatalf("getUserToken() error = %v", err)
	}
	if token != "tok123" {
		t.Errorf("getUserToken() = %q, want %q", token, "tok123")
	}
}

func TestGetUserToken_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := newClientWithBaseURL("apikey", "", server.URL)
	_, err := c.getUserToken("user", "wrong")
	if err == nil {
		t.Error("getUserToken() expected error for 401, got nil")
	}
}

func TestUserPath(t *testing.T) {
	c := newClientWithBaseURL("key", "mytoken", "http://example.com")
	got := c.userPath("/sets/")
	want := "/users/mytoken/sets/"
	if got != want {
		t.Errorf("userPath() = %q, want %q", got, want)
	}
}

func TestFetchAllPagesPagination(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}
	type pageResp struct {
		Count   int    `json:"count"`
		Next    string `json:"next"`
		Results []item `json:"results"`
	}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(pageResp{Count: 2, Results: []item{{ID: 2}}})
		} else {
			_ = json.NewEncoder(w).Encode(pageResp{Count: 2, Next: serverURL + "/?page=2", Results: []item{{ID: 1}}})
		}
	}))
	defer server.Close()
	serverURL = server.URL

	c := newClientWithBaseURL("key", "", server.URL)
	count, results, err := fetchAllPages[struct{ ID int `json:"id"` }](c.http, "/")
	if err != nil {
		t.Fatalf("fetchAllPages() error = %v", err)
	}
	if count != 2 {
		t.Errorf("fetchAllPages() count = %d, want 2", count)
	}
	if len(results) != 2 {
		t.Errorf("fetchAllPages() len(results) = %d, want 2", len(results))
	}
}

func TestGetUserToken_SetsAuthToken(t *testing.T) {
	type tokenResp struct {
		UserToken string `json:"user_token"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(tokenResp{UserToken: "mytoken"})
	}))
	defer server.Close()

	c := newClientWithBaseURL("apikey", "", server.URL)
	token, err := c.getUserToken("user", "pass")
	if err != nil {
		t.Fatalf("getUserToken() error = %v", err)
	}

	authedClient := newClientWithBaseURL("apikey", token, server.URL)
	path := authedClient.userPath("/sets/")
	if path != "/users/mytoken/sets/" {
		t.Errorf("userPath() = %q, want %q", path, "/users/mytoken/sets/")
	}
}

func TestGetLegoSets(t *testing.T) {
	tests := []struct {
		name       string
		response   LegoSetsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns sets", LegoSetsResponse{Count: 1, Results: []Set{{SetNum: "10497-1", Name: "Galaxy Explorer"}}}, 200, false},
		{"server error", LegoSetsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoSets()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoSets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoSets() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoSet(t *testing.T) {
	tests := []struct {
		name       string
		response   Set
		statusCode int
		wantErr    bool
	}{
		{"returns set", Set{SetNum: "10497-1", Name: "Galaxy Explorer", Year: 2022}, 200, false},
		{"not found", Set{}, 404, true},
		{"server error", Set{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoSet("10497-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoSet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.SetNum != tt.response.SetNum {
				t.Errorf("GetLegoSet() set_num = %v, want %v", result.SetNum, tt.response.SetNum)
			}
		})
	}
}

func TestGetLegoSetAlternates(t *testing.T) {
	tests := []struct {
		name       string
		response   LegoSetsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns alternates", LegoSetsResponse{Count: 1, Results: []Set{{SetNum: "moc-1234"}}}, 200, false},
		{"server error", LegoSetsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoSetAlternates("10497-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoSetAlternates() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoSetAlternates() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoSetMinifigs(t *testing.T) {
	tests := []struct {
		name       string
		response   SetMinifigsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns minifigs", SetMinifigsResponse{Count: 2, Results: []SetMinifig{{SetNum: "fig-001", Name: "Astronaut"}}}, 200, false},
		{"server error", SetMinifigsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoSetMinifigs("10497-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoSetMinifigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoSetMinifigs() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoSetParts(t *testing.T) {
	tests := []struct {
		name       string
		response   SetPartsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns parts", SetPartsResponse{Count: 3, Results: []SetPart{{Quantity: 2, Part: Part{PartNum: "3001"}}}}, 200, false},
		{"server error", SetPartsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoSetParts("10497-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoSetParts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoSetParts() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoSetSets(t *testing.T) {
	tests := []struct {
		name       string
		response   LegoSetsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns sub-sets", LegoSetsResponse{Count: 1, Results: []Set{{SetNum: "75192-1"}}}, 200, false},
		{"server error", LegoSetsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoSetSets("10497-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoSetSets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoSetSets() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoColors(t *testing.T) {
	tests := []struct {
		name       string
		response   ColorsResponse
		statusCode int
		wantErr    bool
	}{
		{
			"returns colors",
			ColorsResponse{Count: 2, Results: []PartColor{
				{ID: 0, Name: "Black", RGB: "05131D", IsTrans: false},
				{ID: 1, Name: "Blue", RGB: "0055BF", IsTrans: false},
			}},
			200, false,
		},
		{"server error", ColorsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoColors()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoColors() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoColors() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoColor(t *testing.T) {
	tests := []struct {
		name       string
		response   PartColor
		statusCode int
		wantErr    bool
	}{
		{"returns color", PartColor{ID: 0, Name: "Black", RGB: "05131D", IsTrans: false}, 200, false},
		{"not found", PartColor{}, 404, true},
		{"server error", PartColor{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoColor("0")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.ID != tt.response.ID {
				t.Errorf("GetLegoColor() id = %v, want %v", result.ID, tt.response.ID)
			}
		})
	}
}

func TestGetLegoElement(t *testing.T) {
	tests := []struct {
		name       string
		response   Element
		statusCode int
		wantErr    bool
	}{
		{
			"returns element",
			Element{
				ElementID: "4119739",
				Part:      Part{PartNum: "3001", Name: "Brick 2 x 4"},
				Color:     PartColor{ID: 1, Name: "Blue"},
				DesignID:  "3001",
			},
			200, false,
		},
		{"not found", Element{}, 404, true},
		{"server error", Element{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoElement("4119739")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoElement() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.ElementID != tt.response.ElementID {
				t.Errorf("GetLegoElement() element_id = %v, want %v", result.ElementID, tt.response.ElementID)
			}
		})
	}
}

func TestGetLegoMinifigs(t *testing.T) {
	tests := []struct {
		name       string
		response   MinifigsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns minifigs", MinifigsResponse{Count: 1, Results: []Minifig{{SetNum: "fig-000001", Name: "Spaceman"}}}, 200, false},
		{"server error", MinifigsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoMinifigs()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoMinifigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoMinifigs() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoMinifig(t *testing.T) {
	tests := []struct {
		name       string
		response   Minifig
		statusCode int
		wantErr    bool
	}{
		{"returns minifig", Minifig{SetNum: "fig-000001", Name: "Spaceman", NumParts: 4}, 200, false},
		{"not found", Minifig{}, 404, true},
		{"server error", Minifig{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoMinifig("fig-000001")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoMinifig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.SetNum != tt.response.SetNum {
				t.Errorf("GetLegoMinifig() set_num = %v, want %v", result.SetNum, tt.response.SetNum)
			}
		})
	}
}

func TestGetLegoMinifigParts(t *testing.T) {
	tests := []struct {
		name       string
		response   SetPartsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns parts", SetPartsResponse{Count: 2, Results: []SetPart{{Quantity: 1, Part: Part{PartNum: "3001"}}}}, 200, false},
		{"server error", SetPartsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoMinifigParts("fig-000001")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoMinifigParts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoMinifigParts() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoMinifigSets(t *testing.T) {
	tests := []struct {
		name       string
		response   LegoSetsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns sets", LegoSetsResponse{Count: 1, Results: []Set{{SetNum: "10497-1", Name: "Galaxy Explorer"}}}, 200, false},
		{"server error", LegoSetsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoMinifigSets("fig-000001")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoMinifigSets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoMinifigSets() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoPartCategories(t *testing.T) {
	tests := []struct {
		name       string
		response   PartCategoriesResponse
		statusCode int
		wantErr    bool
	}{
		{"returns categories", PartCategoriesResponse{Count: 1, Results: []PartCategory{{ID: 1, Name: "Baseplates", PartCount: 243}}}, 200, false},
		{"server error", PartCategoriesResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoPartCategories()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoPartCategories() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoPartCategories() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoPartCategory(t *testing.T) {
	tests := []struct {
		name       string
		response   PartCategory
		statusCode int
		wantErr    bool
	}{
		{"returns category", PartCategory{ID: 1, Name: "Baseplates", PartCount: 243}, 200, false},
		{"not found", PartCategory{}, 404, true},
		{"server error", PartCategory{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoPartCategory("1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoPartCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.ID != tt.response.ID {
				t.Errorf("GetLegoPartCategory() id = %v, want %v", result.ID, tt.response.ID)
			}
		})
	}
}

func TestGetLegoThemes(t *testing.T) {
	tests := []struct {
		name       string
		response   ThemesResponse
		statusCode int
		wantErr    bool
	}{
		{"returns themes", ThemesResponse{Count: 1, Results: []Theme{{ID: 1, Name: "Technic"}}}, 200, false},
		{"server error", ThemesResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoThemes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoThemes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoThemes() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoTheme(t *testing.T) {
	tests := []struct {
		name       string
		response   Theme
		statusCode int
		wantErr    bool
	}{
		{"returns theme", Theme{ID: 1, Name: "Technic"}, 200, false},
		{"not found", Theme{}, 404, true},
		{"server error", Theme{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoTheme("1")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoTheme() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.ID != tt.response.ID {
				t.Errorf("GetLegoTheme() id = %v, want %v", result.ID, tt.response.ID)
			}
		})
	}
}

func TestGetLegoSetsPagination(t *testing.T) {
	page1Set := Set{SetNum: "10497-1", Name: "Galaxy Explorer"}
	page2Set := Set{SetNum: "75192-1", Name: "Millennium Falcon"}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(LegoSetsResponse{Count: 2, Results: []Set{page2Set}})
		} else {
			_ = json.NewEncoder(w).Encode(LegoSetsResponse{Count: 2, Next: serverURL + "/?page=2", Results: []Set{page1Set}})
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := newClientWithBaseURL("key", "", server.URL)
	result, err := client.GetLegoSets()
	if err != nil {
		t.Fatalf("GetLegoSets() unexpected error: %v", err)
	}
	if result.Count != 2 {
		t.Errorf("GetLegoSets() count = %d, want 2", result.Count)
	}
	if len(result.Results) != 2 {
		t.Errorf("GetLegoSets() len(results) = %d, want 2", len(result.Results))
	}
	if result.Results[0].SetNum != page1Set.SetNum {
		t.Errorf("GetLegoSets() results[0].SetNum = %q, want %q", result.Results[0].SetNum, page1Set.SetNum)
	}
	if result.Results[1].SetNum != page2Set.SetNum {
		t.Errorf("GetLegoSets() results[1].SetNum = %q, want %q", result.Results[1].SetNum, page2Set.SetNum)
	}
}

func TestGetLegoSetsPaginationErrorOnSecondPage(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(LegoSetsResponse{Count: 2, Next: serverURL + "/?page=2", Results: []Set{{SetNum: "10497-1"}}})
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := newClientWithBaseURL("key", "", server.URL)
	_, err := client.GetLegoSets()
	if err == nil {
		t.Error("GetLegoSets() expected error on second page, got nil")
	}
}

func TestGetLegoParts(t *testing.T) {
	tests := []struct {
		name       string
		response   LegoPartsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns parts", LegoPartsResponse{Count: 1, Results: []PartDetail{{PartNum: "3001", Name: "Brick 2 x 4"}}}, 200, false},
		{"server error", LegoPartsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoParts(PartsFilter{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoParts() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoParts() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoPartsAppliesFilters(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LegoPartsResponse{Count: 0})
	}))
	defer server.Close()

	client := newClientWithBaseURL("key", "", server.URL)
	_, err := client.GetLegoParts(PartsFilter{PartCatID: "5", ColorID: "4", Search: "brick"})
	if err != nil {
		t.Fatalf("GetLegoParts() unexpected error: %v", err)
	}
	for _, want := range []string{"part_cat_id=5", "color_id=4", "search=brick"} {
		if !strings.Contains(capturedQuery, want) {
			t.Errorf("GetLegoParts() query = %q, missing %q", capturedQuery, want)
		}
	}
}

func TestGetLegoPartsPagination(t *testing.T) {
	page1 := PartDetail{PartNum: "3001"}
	page2 := PartDetail{PartNum: "3002"}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(LegoPartsResponse{Count: 2, Results: []PartDetail{page2}})
		} else {
			_ = json.NewEncoder(w).Encode(LegoPartsResponse{Count: 2, Next: serverURL + "/?page=2", Results: []PartDetail{page1}})
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := newClientWithBaseURL("key", "", server.URL)
	result, err := client.GetLegoParts(PartsFilter{})
	if err != nil {
		t.Fatalf("GetLegoParts() unexpected error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("GetLegoParts() len = %d, want 2", len(result.Results))
	}
	if result.Results[0].PartNum != page1.PartNum || result.Results[1].PartNum != page2.PartNum {
		t.Errorf("GetLegoParts() pagination order wrong: %+v", result.Results)
	}
}

func TestGetLegoPart(t *testing.T) {
	tests := []struct {
		name       string
		response   PartDetail
		statusCode int
		wantErr    bool
	}{
		{"returns part", PartDetail{PartNum: "3001", Name: "Brick 2 x 4"}, 200, false},
		{"not found", PartDetail{}, 404, true},
		{"server error", PartDetail{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoPart("3001")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoPart() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.PartNum != tt.response.PartNum {
				t.Errorf("GetLegoPart() part_num = %v, want %v", result.PartNum, tt.response.PartNum)
			}
		})
	}
}

func TestGetLegoPartColors(t *testing.T) {
	tests := []struct {
		name       string
		response   PartColorsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns colors", PartColorsResponse{Count: 1, Results: []PartColorDetail{{ColorID: 4, ColorName: "Red"}}}, 200, false},
		{"server error", PartColorsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoPartColors("3001")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoPartColors() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoPartColors() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoPartColorsPagination(t *testing.T) {
	page1 := PartColorDetail{ColorID: 4, ColorName: "Red"}
	page2 := PartColorDetail{ColorID: 5, ColorName: "Blue"}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(PartColorsResponse{Count: 2, Results: []PartColorDetail{page2}})
		} else {
			_ = json.NewEncoder(w).Encode(PartColorsResponse{Count: 2, Next: serverURL + "/?page=2", Results: []PartColorDetail{page1}})
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := newClientWithBaseURL("key", "", server.URL)
	result, err := client.GetLegoPartColors("3001")
	if err != nil {
		t.Fatalf("GetLegoPartColors() unexpected error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("GetLegoPartColors() len = %d, want 2", len(result.Results))
	}
}

func TestGetLegoPartColor(t *testing.T) {
	tests := []struct {
		name       string
		response   PartColorDetail
		statusCode int
		wantErr    bool
	}{
		{"returns combination", PartColorDetail{ColorID: 4, ColorName: "Red", NumSets: 12}, 200, false},
		{"not found", PartColorDetail{}, 404, true},
		{"server error", PartColorDetail{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoPartColor("3001", "4")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoPartColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.ColorName != tt.response.ColorName {
				t.Errorf("GetLegoPartColor() color_name = %v, want %v", result.ColorName, tt.response.ColorName)
			}
		})
	}
}

func TestGetLegoPartColorSets(t *testing.T) {
	tests := []struct {
		name       string
		response   LegoSetsResponse
		statusCode int
		wantErr    bool
	}{
		{"returns sets", LegoSetsResponse{Count: 1, Results: []Set{{SetNum: "10497-1"}}}, 200, false},
		{"server error", LegoSetsResponse{}, 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()
			client := newClientWithBaseURL("key", "", server.URL)
			result, err := client.GetLegoPartColorSets("3001", "4")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLegoPartColorSets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result.Count != tt.response.Count {
				t.Errorf("GetLegoPartColorSets() count = %v, want %v", result.Count, tt.response.Count)
			}
		})
	}
}

func TestGetLegoPartColorSetsPagination(t *testing.T) {
	page1 := Set{SetNum: "10497-1"}
	page2 := Set{SetNum: "75192-1"}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(LegoSetsResponse{Count: 2, Results: []Set{page2}})
		} else {
			_ = json.NewEncoder(w).Encode(LegoSetsResponse{Count: 2, Next: serverURL + "/?page=2", Results: []Set{page1}})
		}
	}))
	defer server.Close()
	serverURL = server.URL

	client := newClientWithBaseURL("key", "", server.URL)
	result, err := client.GetLegoPartColorSets("3001", "4")
	if err != nil {
		t.Fatalf("GetLegoPartColorSets() unexpected error: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("GetLegoPartColorSets() len = %d, want 2", len(result.Results))
	}
}
