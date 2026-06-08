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

func TestNewAuthenticatedClient_Success(t *testing.T) {
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

func TestNewAuthenticatedClient_Failure(t *testing.T) {
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
