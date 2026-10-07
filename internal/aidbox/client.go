package aidbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var ErrLicenseNotFound = errors.New("Aidbox license not found")

type HTTPClient struct {
	Endpoint string
	Token    string
	Client   *http.Client
}

type Creator struct {
	ID           string `yaml:"id"`
	ResourceType string `yaml:"resourceType"`
}

type Project struct {
	ID           string `yaml:"id"`
	ResourceType string `yaml:"resourceType"`
}

type Info struct {
	Hosting string `yaml:"hosting"`
}

type Meta struct {
	LastUpdated string `yaml:"lastUpdated"`
	CreatedAt   string `yaml:"createdAt"`
	VersionID   string `yaml:"versionId"`
}

type Additional struct {
	ExpirationDays int     `yaml:"expiration-days"`
	BoxURL         *string `yaml:"box-url"`
}

type License struct {
	ID           string     `yaml:"id"`
	Name         string     `yaml:"name"`
	Product      string     `yaml:"product"`
	Type         string     `yaml:"type"`
	Expiration   string     `yaml:"expiration"`
	Status       string     `yaml:"status"`
	MaxInstances int        `yaml:"max-instances"`
	Creator      Creator    `yaml:"creator"`
	Project      Project    `yaml:"project"`
	Offline      bool       `yaml:"offline"`
	Created      string     `yaml:"created"`
	Meta         Meta       `yaml:"meta"`
	Issuer       string     `yaml:"issuer"`
	Info         Info       `yaml:"info"`
	Additional   Additional `yaml:"additional"`
}

// LicenseResponse includes the License and JWT token.
type LicenseResponse struct {
	License License
	JWT     string
}

// APIResponse maps the YAML response from the Aidbox API.
type APIResponse struct {
	Result struct {
		License License `yaml:"license"`
		JWT     string  `yaml:"jwt"`
	} `yaml:"result"`
}

func NewClient(endpoint, token string) *HTTPClient {
	return &HTTPClient{
		Endpoint: endpoint,
		Token:    token,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPClient) CreateLicense(ctx context.Context, name, product, licenseType string) (LicenseResponse, error) {
	params := map[string]interface{}{
		"token":   c.Token,
		"name":    name,
		"product": product,
		"type":    licenseType,
	}

	bodyBytes, err := c.makeAPICall(ctx, "portal.portal/issue-license", params)
	if err != nil {
		return LicenseResponse{}, err
	}

	apiResp, parseErr := parseYAMLResponse(bodyBytes)
	if parseErr != nil {
		return LicenseResponse{}, parseErr
	}
	if apiResp.License.ID == "" {
		return LicenseResponse{}, fmt.Errorf("Aidbox returned a license response without an ID")
	}

	return apiResp, nil
}

func (c *HTTPClient) GetLicense(ctx context.Context, licenseID string) (LicenseResponse, error) {
	params := map[string]interface{}{
		"token": c.Token,
		"id":    licenseID,
	}

	bodyBytes, err := c.makeAPICall(ctx, "portal.portal/get-license", params)
	if err != nil {
		return LicenseResponse{}, err
	}

	apiResp, parseErr := parseYAMLResponse(bodyBytes)
	if parseErr != nil {
		return LicenseResponse{}, parseErr
	}

	if apiResp.License.ID == "" {
		return LicenseResponse{}, fmt.Errorf("Aidbox returned a license response without an ID")
	}
	return apiResp, nil
}

func (c *HTTPClient) DeleteLicense(ctx context.Context, licenseID string) error {
	_, err := c.makeAPICall(ctx, "portal.portal/remove-license", map[string]interface{}{
		"token": c.Token,
		"id":    licenseID,
	})
	return err
}

func (c *HTTPClient) makeAPICall(ctx context.Context, method string, params map[string]interface{}) ([]byte, error) {
	requestBody := map[string]interface{}{
		"method": method,
		"params": params,
	}

	yamlData, err := yaml.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create YAML request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, strings.NewReader(string(yamlData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "text/yaml")
	req.Header.Set("Accept", "text/yaml")

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if resp.StatusCode == http.StatusNotFound && method != "portal.portal/issue-license" {
			return nil, ErrLicenseNotFound
		}
		return nil, fmt.Errorf("Aidbox API returned HTTP status %s", resp.Status)
	}

	return bodyBytes, nil
}

func parseYAMLResponse(bodyBytes []byte) (LicenseResponse, error) {
	var apiResp APIResponse
	if err := yaml.Unmarshal(bodyBytes, &apiResp); err != nil {
		return LicenseResponse{}, fmt.Errorf("failed to parse YAML response: %w", err)
	}
	return LicenseResponse{
		License: apiResp.Result.License,
		JWT:     apiResp.Result.JWT,
	}, nil
}
