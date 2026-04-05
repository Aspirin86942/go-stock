package data

import (
	"testing"

	"go-stock/internal/testenv"
)

func requireIntegrationTest(t *testing.T) {
	testenv.RequireExternalTest(t)
}
