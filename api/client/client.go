package client

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/danielkrainas/tinkersnest/api/v1"
)

const CLIENT_USER_AGENT = "tinkersnest-client/1.0.0"

type Client struct {
	setup      sync.Once
	urlBuilder *v1.URLBuilder
	Endpoint   string
	HTTPClient *http.Client
	UserAgent  string
	AuthToken  AuthToken
}

func New(endpoint string, httpClient *http.Client) *Client {
	return &Client{
		Endpoint:   endpoint,
		HTTPClient: httpClient,
		UserAgent:  CLIENT_USER_AGENT,
	}
}

func (c *Client) urls() *v1.URLBuilder {
	c.setup.Do(func() {
		ub, err := v1.NewURLBuilderFromString(c.Endpoint, false)
		if err != nil {
			panic(fmt.Sprintf("error creating v1 url builder: %v", err))
		}

		c.urlBuilder = ub
	})

	return c.urlBuilder
}

func (c *Client) Ping() error {
	destUrl, err := c.urls().BuildBaseURL()
	if err != nil {
		return err
	}

	r, err := http.NewRequest(http.MethodGet, destUrl, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(r)
	if err != nil {
		return err
	}

	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status returned: %s", resp.Status)
	}

	return nil
}

func (c *Client) do(r *http.Request) (*http.Response, error) {
	r.Header.Add("USER-AGENT", c.UserAgent)
	if c.AuthToken != InvalidToken {
		r.Header.Add("Authorization", "Bearer "+string(c.AuthToken))
	}

	return c.HTTPClient.Do(r)
}
