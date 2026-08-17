package ups

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hiscaler/ups-go/config"
)

var (
	client                   *Client
	ctx                      = context.Background()
	integrationAccountNumber string
	integrationEnv           string
)

func TestMain(m *testing.M) {
	b, err := os.ReadFile("./config/config.json")
	if err != nil {
		fmt.Printf("integration client disabled: %s\n", err.Error())
		os.Exit(m.Run())
	}
	var cfg config.Config
	if err = json.Unmarshal(b, &cfg); err != nil {
		panic(fmt.Sprintf("Parse config file error: %s", err.Error()))
	}
	if cfg.ClientID != "" && cfg.ClientID != "YOUR_CLIENT_ID" {
		client = NewClient(ctx, cfg)
		integrationAccountNumber = cfg.AccountNumber
		integrationEnv = cfg.Env
	} else {
		fmt.Println("integration client disabled: placeholder credentials")
	}
	os.Exit(m.Run())
}
