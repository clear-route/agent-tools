package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const baseURL = "https://graph.microsoft.com/v1.0"

type UserProfile struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

type Client struct {
	cred   azcore.TokenCredential
	scopes []string
	http   *http.Client

	meOnce    sync.Once
	meProfile UserProfile
	meErr     error
}

func NewClient(cred azcore.TokenCredential, scopes []string) *Client {
	return &Client{
		cred:   cred,
		scopes: scopes,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Me() (UserProfile, error) {
	c.meOnce.Do(func() {
		var me struct {
			DisplayName       string `json:"displayName"`
			Mail              string `json:"mail"`
			UserPrincipalName string `json:"userPrincipalName"`
		}
		c.meErr = c.GetJSON(context.Background(), "/me?$select=displayName,mail,userPrincipalName", &me)
		if c.meErr != nil {
			return
		}
		email := me.Mail
		if email == "" {
			email = me.UserPrincipalName
		}
		c.meProfile = UserProfile{DisplayName: me.DisplayName, Email: email}
	})
	return c.meProfile, c.meErr
}

func (c *Client) token(ctx context.Context) (string, error) {
	tok, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: c.scopes})
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}
	return tok.Token, nil
}

func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := path
	if !strings.HasPrefix(path, "http") {
		url = baseURL + path
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	tok, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	return c.http.Do(req)
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

func (c *Client) GetJSON(ctx context.Context, path string, target any) error {
	resp, err := c.Get(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: %s — %s", path, resp.Status, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) GetBody(ctx context.Context, path string) ([]byte, error) {
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s: %s — %s", path, resp.Status, string(b))
	}
	return io.ReadAll(resp.Body)
}
