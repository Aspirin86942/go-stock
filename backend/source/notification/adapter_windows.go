//go:build windows || darwin

package notification

import "go-stock/backend/data"

func sendLocalNotification(appID, title, content, icon string) bool {
	return data.NewAlertWindowsApi(appID, title, content, icon).SendNotification()
}
