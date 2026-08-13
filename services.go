package ups

import (
	"log/slog"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hiscaler/ups-go/config"
)

type service struct {
	config     *config.Config // Config
	logger     *slog.Logger   // Logger
	httpClient *resty.Client  // HTTP client

	tokenMu        sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

// API Services
type services struct {
	Auth     authService     // 认证服务
	Shipment shipmentService // 发货服务
}
