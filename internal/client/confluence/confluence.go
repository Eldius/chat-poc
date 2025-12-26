package confluence

import (
	"bytes"
	"chat-poc/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sessionFilePath = ".db/session.json"
	dataFolder      = ".db"
)

type Options struct {
	BaseUrl         string
	AuthURL         string
	clientID        string
	clientSecret    string
	scopes          []string
	redirectURL     string
	refreshTokenURL string
	responseType    string
	audience        string
	prompt          string
	c               *http.Client
	currSession     RefreshTokenResponse
}

func defaultOptions() *Options {
	return &Options{
		AuthURL:         "https://auth.atlassian.com/authorize",
		BaseUrl:         "https://confluence.atlassian.com",
		clientID:        "1234567890",
		clientSecret:    "secret",
		scopes:          config.ConfluenceScopesProp.Value.([]string),
		redirectURL:     "http://localhost:8080/auth/callback",
		responseType:    "code",
		c:               http.DefaultClient,
		audience:        "api.atlassian.com",
		prompt:          "consent",
		refreshTokenURL: "https://auth.atlassian.com/oauth/token",
	}
}

type ClientOption func(*Options)

type refreshTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectUri  string `json:"redirect_uri"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func WithBaseUrl(url string) ClientOption {
	return func(o *Options) {
		o.BaseUrl = url
	}
}

func WithAuthURL(url string) ClientOption {
	return func(o *Options) {
		o.AuthURL = url
	}
}

func WithClientID(id string) ClientOption {
	return func(o *Options) {
		o.clientID = id
	}
}

func WithScopes(scopes []string) ClientOption {
	return func(o *Options) {
		o.scopes = scopes
	}
}

func WithRedirectURL(url string) ClientOption {
	return func(o *Options) {
		o.redirectURL = url
	}
}

func WithResponseType(responseType string) ClientOption {
	return func(o *Options) {
		o.responseType = responseType
	}
}

func WithHTTPClient(c *http.Client) ClientOption {
	return func(o *Options) {
		o.c = c
	}
}

func WithAudience(audience string) ClientOption {
	return func(o *Options) {
		o.audience = audience
	}
}

func WithPrompt(prompt string) ClientOption {
	return func(o *Options) {
		o.prompt = prompt
	}
}

func WithRefreshTokenURL(url string) ClientOption {
	return func(o *Options) {
		o.refreshTokenURL = url
	}
}

func WithSession(session RefreshTokenResponse) ClientOption {
	return func(o *Options) {
		o.currSession = session
	}
}

func WithClientSecret(secret string) ClientOption {
	return func(o *Options) {
		o.clientSecret = secret
	}
}

type Client struct {
	opts *Options
}

func NewClient(opts ...ClientOption) *Client {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return &Client{opts: options}
}

func (c *Client) Authenticate(ctx context.Context) error {
	fmt.Println("Starting authentication...")
	u, err := url.Parse(c.opts.AuthURL)
	if err != nil {
		return fmt.Errorf("parsing auth url: %w", err)
	}
	q := u.Query()
	q.Set("client_id", c.opts.clientID)
	q.Set("redirect_uri", c.opts.redirectURL)
	q.Set("response_type", c.opts.responseType)
	q.Set("scope", strings.Join(c.opts.scopes, " "))
	q.Set("state", strconv.Itoa(rand.Int()))
	q.Set("audience", c.opts.audience)
	q.Set("prompt", "")

	u.RawQuery = q.Encode()

	redirectURL, err := url.Parse(c.opts.redirectURL)
	if err != nil {
		return fmt.Errorf("parsing redirect url: %w", err)
	}

	var wg sync.WaitGroup

	wg.Go(func() {

		fmt.Println("Starting auth service on port", redirectURL.Port())
		s := http.Server{Addr: fmt.Sprintf(":%s", redirectURL.Port())}

		mux := http.NewServeMux()
		mux.HandleFunc(http.MethodGet+" "+redirectURL.Path, func(w http.ResponseWriter, r *http.Request) {
			defer serverShutdown(ctx, &s)

			_, _ = fmt.Fprintf(w, "Authentication successful!")
			_, _ = fmt.Fprintf(w, "%#v", r.URL.Query())

			_, _ = fmt.Println("Authentication successful!")
			_, _ = fmt.Printf("%#v\n---\n", r.URL.Query())

			b, err := json.Marshal(refreshTokenRequest{
				GrantType:    "authorization_code",
				ClientId:     c.opts.clientID,
				ClientSecret: c.opts.clientSecret,
				Code:         r.URL.Query().Get("code"),
				RedirectUri:  c.opts.redirectURL,
			})
			if err != nil {
				_, _ = fmt.Fprintf(w, "Error: %s", err)
				return
			}
			req, err := http.NewRequest(http.MethodPost, c.opts.refreshTokenURL, bytes.NewBuffer(b))
			if err != nil {
				_, _ = fmt.Fprintf(w, "Error: %s", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			res, err := c.opts.c.Do(req.WithContext(ctx))
			if err != nil {
				_, _ = fmt.Fprintf(w, "Error: %s", err)
				fmt.Printf("error: %#v\n", res)
				return
			}
			defer func() {
				_ = res.Body.Close()
			}()

			b, err = io.ReadAll(res.Body)
			_, _ = fmt.Fprintf(w, "%s", b)
			fmt.Printf("body: %s\n", b)

			_ = json.Unmarshal(b, &c.opts.currSession)
			c.PersistSession(c.opts.currSession)

			http.Redirect(w, r, "/shutdown", http.StatusTemporaryRedirect)
		})
		mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, "Shutting down server...")
			go serverShutdown(ctx, &s)
		})
		s.Handler = mux
		if err := s.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	})

	fmt.Printf("---\nPlease, open this link on your browser and log in: %s\n---\n", u.String())
	fmt.Println("Waiting for authentication...")

	wg.Wait()

	return nil
}

func serverShutdown(ctx context.Context, s *http.Server) {
	fmt.Println("Shutting down server...")
	time.Sleep(15 * time.Second)
	_ = s.Shutdown(ctx)
}

func (c *Client) PersistSession(s RefreshTokenResponse) {
	stat, err := os.Stat(dataFolder)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dataFolder, 0600)
			if err != nil {
				fmt.Println("Error creating session directory:", err)
				return
			}
		}
	}
	if stat == nil || !stat.IsDir() {
		fmt.Println("Invalid session file")
	}

	f, err := os.OpenFile(sessionFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("Error creating session file:", err)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	b, _ := json.MarshalIndent(s, "", "  ")
	_, _ = f.Write(b)
}

func (c *Client) LoadSession() error {

	stat, err := os.Stat(sessionFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No session file found")
			return nil
		}
	}
	if stat == nil || stat.IsDir() {
		return fmt.Errorf("invalid session file")
	}
	f, err := os.Open(sessionFilePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return json.NewDecoder(f).Decode(&c.opts.currSession)
}

func (c *Client) List() string {

	return ""
}

type APIRateLimitsInfo struct {
	// RateLimit
	// X-RateLimit-Limit: The maximum request rate enforced for the current rate-limit scope. For burst rate limits, this reflects the allowed requests per second.
	RateLimit int64

	// Remaining
	// X-RateLimit-Remaining: The remaining request capacity within the current rate-limit window. For burst rate limits, this represents remaining requests in the current second.
	Remaining int64

	// Reset
	// X-RateLimit-Reset: ISO 8601 timestamp when the current window resets
	Reset int64

	// NearLimit
	// X-RateLimit-NearLimit: Returns true when less than 20% of the quota remains
	NearLimit bool

	// Reason
	// RateLimit-Reason	The reason for throttling:
	// • confluence-quota-global-based – Global Pool limits breached
	// • confluence-quota-tenant-based – Per-Tenant Pool limits breached
	Reason string

	// RetryAfter
	// Retry-After: Only returned with 429 responses. Indicates how many seconds to wait before retrying
	RetryAfter int64
}
