package ups

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/hiscaler/ups-go/entity"
)

// 认证服务
type authService struct {
	*service
}

// Token 显式获取 OAuth access token（会更新本地缓存）
func (s authService) Token(ctx context.Context) (entity.Token, error) {
	return s.fetchToken(ctx)
}

// ensureAccessToken 确保本地有未过期的 access token
func (s *service) ensureAccessToken(ctx context.Context) (string, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	if s.accessToken != "" && time.Now().Before(s.tokenExpiresAt.Add(-tokenSkew)) {
		return s.accessToken, nil
	}

	token, err := s.doFetchToken(ctx)
	if err != nil {
		return "", err
	}
	s.cacheToken(token)
	return token.AccessToken, nil
}

// invalidateToken 清空本地 token 缓存
func (s *service) invalidateToken() {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	s.accessToken = ""
	s.tokenExpiresAt = time.Time{}
}

// fetchToken 拉取并缓存 token
func (s authService) fetchToken(ctx context.Context) (entity.Token, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	token, err := s.doFetchToken(ctx)
	if err != nil {
		return entity.Token{}, err
	}
	s.cacheToken(token)
	return token, nil
}

// cacheToken 写入本地 token 与过期时间
func (s *service) cacheToken(token entity.Token) {
	s.accessToken = token.AccessToken
	expiresIn, _ := strconv.ParseInt(token.ExpiresIn, 10, 64)
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	s.tokenExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
}

// doFetchToken 调用 OAuth token 接口
func (s *service) doFetchToken(ctx context.Context) (entity.Token, error) {
	if s.config.ClientID == "" || s.config.ClientSecret == "" {
		return entity.Token{}, errors.New("ClientID 或 ClientSecret 不能为空")
	}

	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetBasicAuth(s.config.ClientID, s.config.ClientSecret).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetFormData(map[string]string{
			"grant_type": "client_credentials",
		}).
		SetHeader("x-merchant-id", s.config.AccountNumber).
		Post("/security/v1/oauth/token")
	if err = recheckError(resp, err); err != nil {
		return entity.Token{}, err
	}

	var token entity.Token
	if err = json.Unmarshal(resp.Body(), &token); err != nil {
		return entity.Token{}, err
	}
	if strings.TrimSpace(token.AccessToken) == "" {
		return entity.Token{}, errors.New("access_token 为空")
	}
	return token, nil
}

// doWithAuth 带 Bearer 鉴权执行请求；遇 401 清缓存并重试一次
func (s *service) doWithAuth(ctx context.Context, call func(ctx context.Context, token string) error) error {
	token, err := s.ensureAccessToken(ctx)
	if err != nil {
		return err
	}
	err = call(ctx, token)
	if err == nil {
		return nil
	}
	if !isUnauthorized(err) {
		return err
	}

	s.invalidateToken()
	token, err = s.ensureAccessToken(ctx)
	if err != nil {
		return err
	}
	return call(ctx, token)
}

// isUnauthorized 判断错误是否为未授权
func isUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "401") ||
		strings.Contains(strings.ToLower(msg), "unauthorized")
}
