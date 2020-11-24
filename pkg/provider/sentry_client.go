package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
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
func (c *SentryClient) Check() error {
	projectsURL, err := c.url.Parse("projects/")
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodGet, projectsURL.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(request)
	if err != nil {
		// return fmt.Errorf("call to %v failed: %v", projectsURL.String(), err)
		return fmt.Errorf("call to %v failed: %v", request, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if _, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error().Err(err).Msg("could not read auth-api response")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("call to %v returned status %v", request, resp.StatusCode)
	}

	return nil
}
