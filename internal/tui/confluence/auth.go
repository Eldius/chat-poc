package confluence

import (
	"chat-poc/internal/client/confluence"
	"chat-poc/internal/config"
	"context"
	"fmt"

	httpclient "github.com/eldius/initial-config-go/http/client"
)

func StartAuth(ctx context.Context) error {
	c := confluence.NewClient(
		confluence.WithAuthURL(config.GetConfluenceAuthURL()),
		confluence.WithResponseType(config.GetConfluenceAuthResponseType()),
		confluence.WithRedirectURL(config.GetConfluenceAuthRedirectURL()),
		confluence.WithClientID(config.GetConfluenceClientID()),
		confluence.WithScopes(config.GetConfluenceScopes()),
		confluence.WithHTTPClient(httpclient.NewHTTPClient()),
		confluence.WithPrompt(config.GetConfluenceAuthPrompt()),
		confluence.WithAudience(config.GetConfluenceAuthAudience()),
		confluence.WithClientSecret(config.GetConfluenceClientSecret()),
	)

	if err := c.LoadSession(); err != nil {
		return fmt.Errorf("loading confluence session: %w", err)
	}

	if err := c.Authenticate(ctx); err != nil {
		return fmt.Errorf("authenticating confluence: %w", err)
	}

	return nil
}
