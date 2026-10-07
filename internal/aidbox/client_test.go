package aidbox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCreateLicense(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected request method %q", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "text/yaml" {
			t.Errorf("unexpected content type %q", got)
		}

		var request struct {
			Method string            `yaml:"method"`
			Params map[string]string `yaml:"params"`
		}
		if err := yaml.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if request.Method != "portal.portal/issue-license" {
			t.Errorf("unexpected API method %q", request.Method)
		}
		if request.Params["token"] != "secret" || request.Params["name"] != "dev" ||
			request.Params["product"] != "aidbox" || request.Params["type"] != "development" {
			t.Errorf("unexpected API parameters %#v", request.Params)
		}

		w.Header().Set("Content-Type", "text/yaml")
		_, _ = w.Write([]byte("result:\n  license:\n    id: license-123\n    name: dev\n    product: aidbox\n    type: development\n    max-instances: 3\n    meta:\n      versionId: v1\n  jwt: signed-license\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret")
	response, err := client.CreateLicense(context.Background(), "dev", "aidbox", "development")
	if err != nil {
		t.Fatalf("CreateLicense returned error: %v", err)
	}
	if response.License.ID != "license-123" || response.License.MaxInstances != 3 ||
		response.License.Meta.VersionID != "v1" || response.JWT != "signed-license" {
		t.Fatalf("unexpected license response: %#v", response)
	}
}

func TestGetLicenseRequiresResponseID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("result:\n  license:\n    name: missing-id\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret")
	_, err := client.GetLicense(context.Background(), "license-123")
	if err == nil || !strings.Contains(err.Error(), "without an ID") {
		t.Fatalf("expected missing ID error, got %v", err)
	}
}

func TestAPIErrorDoesNotExposeResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "sensitive server detail", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret")
	_, err := client.GetLicense(context.Background(), "license-123")
	if err == nil {
		t.Fatal("expected API error")
	}
	if strings.Contains(err.Error(), "sensitive server detail") {
		t.Fatalf("API error exposed response body: %v", err)
	}
	if !strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Fatalf("API error does not identify status: %v", err)
	}
}

func TestGetLicenseNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	client := NewClient(server.URL, "secret")
	_, err := client.GetLicense(context.Background(), "license-123")
	if !errors.Is(err, ErrLicenseNotFound) {
		t.Fatalf("expected ErrLicenseNotFound, got %v", err)
	}
}

func TestDeleteLicense(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Method string            `yaml:"method"`
			Params map[string]string `yaml:"params"`
		}
		if err := yaml.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if request.Method != "portal.portal/remove-license" || request.Params["id"] != "license-123" ||
			request.Params["token"] != "secret" {
			t.Errorf("unexpected delete request %#v", request)
		}
		_, _ = w.Write([]byte("result: removed\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret")
	if err := client.DeleteLicense(context.Background(), "license-123"); err != nil {
		t.Fatalf("DeleteLicense returned error: %v", err)
	}
}
