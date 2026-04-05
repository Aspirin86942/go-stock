//go:build darwin
// +build darwin

package data

import (
	"go-stock/internal/testenv"
	"testing"

	"github.com/go-toast/toast"
)

// @Author 2lovecode
// @Date 2025/02/06 17:50
// @Desc
// -----------------------------------------------------------------------------------

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
		t.Fatalf("send darwin alert: %v", err)
	}
}
