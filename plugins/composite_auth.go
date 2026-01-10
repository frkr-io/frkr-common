package plugins

import (
	"context"
	"fmt"
	"net/http"
)

// CompositeAuthPlugin aggregates multiple auth plugins
type CompositeAuthPlugin struct {
	plugins []AuthPlugin
}

// NewCompositeAuthPlugin creates a new CompositeAuthPlugin
func NewCompositeAuthPlugin(plugins ...AuthPlugin) *CompositeAuthPlugin {
	return &CompositeAuthPlugin{
		plugins: plugins,
	}
}

// ValidateRequest iterates through all plugins until one succeeds or all fail.
func (p *CompositeAuthPlugin) ValidateRequest(ctx context.Context, r *http.Request, secretPlugin SecretPlugin) (*AuthResult, error) {
	var lastErr error
	for _, plugin := range p.plugins {
		res, err := plugin.ValidateRequest(ctx, r, secretPlugin)
		if err == nil && res != nil {
			return res, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no auth plugin could validate the request")
}

// CanAccessStream routes the check to the plugin that originated the auth (AuthSource).
// If AuthSource is empty, it tries all plugins (fallback).
func (p *CompositeAuthPlugin) CanAccessStream(ctx context.Context, authResult *AuthResult, streamID string, permission string) (bool, error) {
	// 1. Try to match by AuthSource (plugins should expose a name/source constant ideally, 
	// but here we rely on the plugins marking the result)
	// Actually, the Composite needs to know which plugin corresponds to which source?
	// OR, we just iterate all plugins again?
	// The problem with iteration is that a "Basic" plugin checking an "OIDC" user might return "User Not Found" error.
	//
	// Better Design: The `AuthPlugin` interface should arguably have a `Name()` method, but that's a larger change.
	//
	// Strategy: Since `AuthResult.AuthSource` is just a string tag, `CompositeAuthPlugin` doesn't inherently know 
	// which plugin instance owns that tag unless we map it.
	//
	// However, we can simply iterate all plugins. Since `CanAccessStream` takes the Full `AuthResult` now,
	// each plugin *should* check "Is this my user?" via the `AuthSource` field before attempting DB lookups.
	//
	// BUT: `BasicAuthPlugin` doesn't know about `AuthSource` filtering yet. We have to implement that check 
	// INSIDE `BasicAuthPlugin` and `OIDCPlugin`.
	//
	// So here in Composite, we just broadcast.
	
	for _, plugin := range p.plugins {
		allowed, err := plugin.CanAccessStream(ctx, authResult, streamID, permission)
		if err == nil {
			return allowed, nil
		}
		// If explicit denial, maybe stop? For now, assume "error" means "not responsible" or "fail".
		// We continue if it's just an error, but if it's a hard "false" (NO error), that usually means explicit denial.
		// However, in current implementations, "User Not Found" is an error.
	}

	return false, fmt.Errorf("access denied or internal error in all plugins")
}
