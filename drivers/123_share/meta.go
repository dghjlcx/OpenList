package _123Share

import (
	"github.com/OpenListTeam/OpenList/v4/internal/driver"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
)

type Addition struct {
    ShareKey string `json:"sharekey" required:"true"`
    SharePwd string `json:"sharepassword"`
    driver.RootID
    AccessToken string `json:"accesstoken" type:"text"`
    Platform    string `json:"platform" type:"string" default:"web" help:"Platform header (e.g., 'android' to bypass limits)"`  // 如果未加，从之前步骤
    UserAgent   string `json:"useragent" type:"text" default:"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) openlist-client" help:"Custom User-Agent for API requests (grab from Android App for bypass)"`  // ← 新增
}

var config = driver.Config{
	Name:        "123PanShare",
	LocalSort:   true,
	NoUpload:    true,
	DefaultRoot: "0",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Pan123Share{}
	})
}
