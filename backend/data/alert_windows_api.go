//go:build windows
// +build windows

package data

import (
	"go-stock/backend/logger"

	"github.com/go-toast/toast"
)

// AlertWindowsApi @Author spark
// @Date 2025/1/8 9:40
// @Desc
// -----------------------------------------------------------------------------------
var alertWindowsLog = dataModuleLogger(logger.SinkApp, "data.alert_windows")

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
		//logger.SugaredLogger.Error("本地推送未开启")
		return false
	}

	notification := toast.Notification{
		AppID:    a.AppID,
		Title:    a.Title,
		Message:  a.Content,
		Icon:     a.Icon,
		Duration: "short",
		Audio:    toast.Default,
	}
	err := notification.Push()
	if err != nil {
		alertWindowsLog.Errorf("data.alert_windows.push_failed", "windows notification push failed: %v", err)
		return false
	}
	return true
}
