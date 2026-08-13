package ups

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-resty/resty/v2"
	"github.com/hiscaler/ups-go/config"
	"github.com/hiscaler/ups-go/entity"
)

const (
	Version   = "0.0.1"
	userAgent = "UPS API Client-Golang/" + Version + " (https://github.com/hiscaler/ups-go)"
)

const (
	ProdBaseUrl = "https://onlinetools.ups.com"
	TestBaseUrl = "https://wwwcie.ups.com"

	defaultVersion        = "v2409"
	defaultLabelVersion   = "v1"
	defaultTransactionSrc = "ups-go"
	tokenSkew             = 60 * time.Second // 提前刷新，避免边界过期
)

// Client UPS API 客户端
type Client struct {
	config     *config.Config // 配置
	httpClient *resty.Client  // Resty Client
	Services   services       // API Services
}

// NewClient 创建 UPS API 客户端
func NewClient(ctx context.Context, cfg config.Config) *Client {
	l := createLogger()
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}
	if cfg.Version == "" {
		cfg.Version = defaultVersion
	}
	if cfg.LabelVersion == "" {
		cfg.LabelVersion = defaultLabelVersion
	}
	if cfg.TransactionSrc == "" {
		cfg.TransactionSrc = defaultTransactionSrc
	}

	upsClient := &Client{
		config: &cfg,
	}
	baseUrl := ProdBaseUrl
	if cfg.Env != entity.Prod {
		baseUrl = TestBaseUrl
	}
	httpClient := resty.New().
		SetDebug(cfg.Debug).
		SetBaseURL(baseUrl).
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
			"User-Agent":   userAgent,
		}).
		SetTimeout(time.Duration(cfg.Timeout) * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)
	upsClient.httpClient = httpClient

	xService := &service{
		config:     &cfg,
		logger:     l.l,
		httpClient: upsClient.httpClient,
	}
	upsClient.Services = services{
		Auth:     authService{service: xService},
		Shipment: shipmentService{service: xService},
	}
	return upsClient
}

// UPSErrorResponse UPS 错误响应
type UPSErrorResponse struct {
	Response struct {
		Errors []UPSError `json:"errors"`
	} `json:"response"`
}

// UPSError UPS 错误项
type UPSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// errorWrap 错误包装
func errorWrap(code, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "未知错误"
	}
	if code == "" {
		return errors.New(message)
	}
	return fmt.Errorf("%s %s", code, message)
}

// invalidInput 将 ozzo 校验错误整理为可读错误
func invalidInput(e error) error {
	var errs validation.Errors
	if !errors.As(e, &errs) {
		return e
	}

	if len(errs) == 0 {
		return nil
	}

	fields := make([]string, 0)
	messages := make([]string, 0)
	for field := range errs {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	for _, field := range fields {
		e1 := errs[field]
		if e1 == nil {
			continue
		}

		var errObj validation.ErrorObject
		if errors.As(e1, &errObj) {
			e1 = errObj
		} else {
			var errs1 validation.Errors
			if errors.As(e1, &errs1) {
				e1 = invalidInput(errs1)
				if e1 == nil {
					continue
				}
			}
		}

		messages = append(messages, e1.Error())
	}
	return errors.New(strings.Join(messages, "; "))
}

// parseUPSErrors 解析 UPS response.errors 错误体
func parseUPSErrors(body []byte) error {
	var errResp UPSErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil
	}
	if len(errResp.Response.Errors) == 0 {
		return nil
	}
	parts := make([]string, 0, len(errResp.Response.Errors))
	for _, e := range errResp.Response.Errors {
		parts = append(parts, fmt.Sprintf("%s %s", e.Code, e.Message))
	}
	return errors.New(strings.Join(parts, "; "))
}

// recheckError 检查 HTTP/业务错误
func recheckError(resp *resty.Response, e error) error {
	if e != nil {
		if errors.Is(e, http.ErrHandlerTimeout) {
			return errorWrap("408", e.Error())
		}
		return e
	}

	if resp == nil {
		return errors.New("空响应")
	}

	if resp.IsError() {
		if err := parseUPSErrors(resp.Body()); err != nil {
			return err
		}
		return fmt.Errorf("%d %s", resp.StatusCode(), strings.TrimSpace(string(resp.Body())))
	}

	if err := parseUPSErrors(resp.Body()); err != nil {
		return err
	}
	return nil
}

// newTransID 生成请求唯一 transId
func newTransID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// apiVersion 返回发货/取消 API 版本
func (s *service) apiVersion() string {
	if s.config.Version == "" {
		return defaultVersion
	}
	return s.config.Version
}

// labelAPIVersion 返回面单补打 API 版本
func (s *service) labelAPIVersion() string {
	if s.config.LabelVersion == "" {
		return defaultLabelVersion
	}
	return s.config.LabelVersion
}

// withAuthHeaders 注入 Bearer、transId、transactionSrc
func (s *service) withAuthHeaders(req *resty.Request, token string) *resty.Request {
	return req.
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("transId", newTransID()).
		SetHeader("transactionSrc", s.config.TransactionSrc)
}
