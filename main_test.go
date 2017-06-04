package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	server := newServer()

	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Error("Failed to ping the server")
	}

	assert.True(t, strings.Contains(w.Body.String(), "pong"), "Ping endpoint returned the wrong result")
}

func TestAuth(t *testing.T) {
	server := newServer()

	req, err := http.NewRequest("GET", "/auth", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != 307 {
		t.Errorf("Failed to call the auth endpoint, returned %d", w.Code)
	}

	assert.True(t, strings.Contains(w.Result().Header.Get("Location"), "https://auth.getmondo.co.uk"), "Ping endpoint returned the wrong result")
}
