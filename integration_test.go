package main

import (
	"testing"

	"go-stock/internal/testenv"
)

func requireIntegrationTest(t *testing.T) {
	t.Helper()
	testenv.RequireExternalTest(t)
}
