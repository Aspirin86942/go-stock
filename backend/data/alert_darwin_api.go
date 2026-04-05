//go:build darwin
// +build darwin

package data

import (
	"fmt"
	"go-stock/backend/logger"

	"os/exec"
)

// AlertWindowsApi @Author 2lovecode
// @Date 2025/02/06 17:50
// @Desc
// -----------------------------------------------------------------------------------
var alertDarwinLog = dataModuleLogger(logger.SinkApp, "data.alert_darwin")

type AlertWindowsApi struct {
	AppID string
	// 窗口标题
	Title string
	// 窗口内容
	Content string
	// 窗口图标
	Icon string
}

func NewAlertWindowsApi(AppID string, Title string, Content string, Icon string) *AlertWindowsApi {
	return &AlertWindowsApi{
		AppID:   AppID,
		Title:   Title,
		Content: Content,
		Icon:    Icon,
	}
}

func (a AlertWindowsApi) SendNotification() bool {
	if GetSettingConfig().LocalPushEnable == false {
		alertDarwinLog.Warn("data.alert_darwin.disabled", "本地推送未开启")
		return false
	}

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, a.Content, a.Title)

	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		alertDarwinLog.Errorf("data.alert_darwin.push_failed", "darwin notification push failed: %v", err)
		return false
	}

	return true
}
