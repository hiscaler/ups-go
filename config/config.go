package config

// Config UPS SDK 配置
type Config struct {
	Debug          bool   `json:"debug"`          // 是否启用调试模式
	Env            string `json:"env"`            // 环境：prod / test / dev
	Timeout        int    `json:"timeout"`        // HTTP 超时设定（单位：秒）
	ClientID       string `json:"clientId"`       // OAuth Client ID
	ClientSecret   string `json:"clientSecret"`   // OAuth Client Secret
	AccountNumber  string `json:"accountNumber"`  // UPS 账户号（ShipperNumber）
	TransactionSrc string `json:"transactionSrc"` // 调用方应用标识
	Version        string `json:"version"`        // API 版本，默认 v2409
	LabelVersion   string `json:"labelVersion"`   // Label Recovery API 版本，默认 v1
}
