package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// SentryClient is a client for the Sentry API.
type SentryClient struct {
	httpClient *http.Client
	token      string
	url        *url.URL
}

// NewSentryClient creates a new SentryClient instance.
func NewSentryClient(urlStr, token string) (*SentryClient, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse URL %#v, %v", urlStr, err)
	}

	return &SentryClient{
		httpClient: &http.Client{},
		token:      token,
		url:        parsedURL,
	}, nil
}

// Check does a test request to check if the client is set up correctly.
func (c *SentryClient) Check(ctx context.Context) error {
	targetURL, err := c.url.Parse("projects/")
	if err != nil {
		return err
	}
	resp, responseBody, err := c.doRequest(ctx, http.MethodGet, targetURL.String(), map[string]string{})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request to %v returned unexpected status %v, response body: %#v", targetURL.String(), resp.StatusCode, string(responseBody))
	}
	return nil
}

// CreateProject creates the project in Sentry.  A project in Sentry is
// identified by the slug, so there is no need to return anything.
func (c *SentryClient) CreateProject(ctx context.Context, organizationSlug, teamSlug, name, slug string) error {
	targetURL, err := c.url.Parse(fmt.Sprintf("teams/%s/%s/projects/", organizationSlug, teamSlug))
	if err != nil {
		return err
	}
	resp, responseBody, err := c.doRequest(ctx, http.MethodPost, targetURL.String(), map[string]string{
		"name": name,
		"slug": slug,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request to %v returned unexpected status %v, response body: %#v", targetURL.String(), resp.StatusCode, string(responseBody))
	}
	return nil
}

// DeleteProject creates the project from Sentry.
func (c *SentryClient) DeleteProject(ctx context.Context, organizationSlug, slug string) error {
	targetURL, err := c.url.Parse(fmt.Sprintf("projects/%s/%s/", organizationSlug, slug))
	if err != nil {
		return err
	}
	resp, responseBody, err := c.doRequest(ctx, http.MethodDelete, targetURL.String(), map[string]string{})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request to %v returned unexpected status %v, response body: %#v", targetURL.String(), resp.StatusCode, string(responseBody))
	}
	return nil
}

func (c *SentryClient) doRequest(ctx context.Context, method, url string, data map[string]string) (*http.Response, []byte, error) {
	serializedData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(serializedData))
	if err != nil {
		return nil, nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "Application/json")

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("call to %s %s failed: %v", request.Method, request.URL.String(), err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := ioutil.ReadAll(resp.Body)
	return resp, responseBody, err
}
