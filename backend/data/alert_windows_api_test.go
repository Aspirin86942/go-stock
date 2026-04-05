//go:build windows
// +build windows

package data

import (
	"go-stock/internal/testenv"
	"testing"

	"github.com/go-toast/toast"
)

// @Author spark
// @Date 2025/1/8 9:40
// @Desc
//-----------------------------------------------------------------------------------

func TestReleaseSmoke_Alert(t *testing.T) {
	testenv.RequireReleaseSmokeTest(t)
	notification := toast.Notification{
		AppID:    "go-stock",
		Title:    "go-stock smoke",
		Message:  "release smoke alert",
		Icon:     "../../build/appicon.png",
		Duration: "short",
		Audio:    toast.Default,
	}
	err := notification.Push()
	if err != nil {
		t.Fatalf("send windows alert: %v", err)
	}
}
