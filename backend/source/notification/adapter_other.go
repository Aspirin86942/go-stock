//go:build !windows && !darwin

package notification

func sendLocalNotification(appID, title, content, icon string) bool {
	return false
}
