package rebrickable

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
