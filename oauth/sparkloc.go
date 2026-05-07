package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func init() {
	Register("sparkloc", &SparklocProvider{})
}

type SparklocProvider struct{}

func (p *SparklocProvider) GetName() string {
	return "Sparkloc"
}

func (p *SparklocProvider) IsEnabled() bool {
	return common.SparklocOAuthEnabled()
}

func (p *SparklocProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	if code == "" {
		return nil, NewOAuthError(i18n.MsgOAuthInvalidCode, nil)
	}

	redirectURI := fmt.Sprintf("%s/oauth/sparkloc", requestOrigin(c))
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, common.SparklocTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	credentials := common.SparklocClientId + ":" + common.SparklocClientSecret
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Sparkloc] ExchangeToken error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Sparkloc"}, err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tokenRes struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		IDToken      string `json:"id_token"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
		Message      string `json:"message"`
	}
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Sparkloc] ExchangeToken decode error: %s", err.Error()))
		return nil, err
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices || tokenRes.AccessToken == "" {
		raw := tokenRes.ErrorDesc
		if raw == "" {
			raw = tokenRes.Message
		}
		if raw == "" {
			raw = string(body)
		}
		if len(raw) > 500 {
			raw = raw[:500] + "..."
		}
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Sparkloc] ExchangeToken failed: status=%d body=%s", res.StatusCode, raw))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthTokenFailed, map[string]any{"Provider": "Sparkloc"}, raw)
	}

	return &OAuthToken{
		AccessToken:  tokenRes.AccessToken,
		TokenType:    tokenRes.TokenType,
		RefreshToken: tokenRes.RefreshToken,
		ExpiresIn:    tokenRes.ExpiresIn,
		Scope:        tokenRes.Scope,
		IDToken:      tokenRes.IDToken,
	}, nil
}

func (p *SparklocProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, common.SparklocUserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Sparkloc] GetUserInfo error: %s", err.Error()))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthConnectFailed, map[string]any{"Provider": "Sparkloc"}, err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		raw := string(body)
		if len(raw) > 500 {
			raw = raw[:500] + "..."
		}
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Sparkloc] GetUserInfo failed: status=%d body=%s", res.StatusCode, raw))
		return nil, NewOAuthErrorWithRaw(i18n.MsgOAuthGetUserErr, map[string]any{"Provider": "Sparkloc"}, raw)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		logger.LogError(ctx, fmt.Sprintf("[OAuth-Sparkloc] GetUserInfo decode error: %s", err.Error()))
		return nil, err
	}

	providerUserID := firstString(raw, "sub", "id", "uid", "user_id", "uuid")
	if providerUserID == "" {
		logger.LogError(ctx, "[OAuth-Sparkloc] GetUserInfo failed: missing user id")
		return nil, NewOAuthError(i18n.MsgOAuthUserInfoEmpty, map[string]any{"Provider": "Sparkloc"})
	}

	username := firstString(raw, "username", "preferred_username", "login", "name")
	displayName := firstString(raw, "display_name", "nickname", "name", "username", "preferred_username")
	email := firstString(raw, "email", "mail")

	if username == "" {
		username = "sparkloc_" + providerUserID
	}

	return &OAuthUser{
		ProviderUserID: providerUserID,
		Username:       sanitizeOAuthUsername(username),
		DisplayName:    displayName,
		Email:          email,
		Extra:          raw,
	}, nil
}

func (p *SparklocProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsSparklocIdAlreadyTaken(providerUserID)
}

func (p *SparklocProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.SparklocId = providerUserID
	return user.FillUserBySparklocId()
}

func (p *SparklocProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.SparklocId = providerUserID
}

func (p *SparklocProvider) GetProviderPrefix() string {
	return "sparkloc_"
}

func requestOrigin(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = c.GetHeader("X-Forwarded-Scheme")
	}
	if scheme == "" {
		scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return strings.TrimRight(scheme+"://"+host, "/")
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case float64:
			return fmt.Sprintf("%.0f", typed)
		case json.Number:
			return typed.String()
		}
	}
	return ""
}

func sanitizeOAuthUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' ||
			r == '-' {
			builder.WriteRune(r)
		}
	}
	cleaned := builder.String()
	if cleaned == "" {
		return ""
	}
	if len(cleaned) > model.UserNameMaxLength {
		cleaned = cleaned[:model.UserNameMaxLength]
	}
	return cleaned
}
