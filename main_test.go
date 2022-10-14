package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverse(t *testing.T) {
	assert.Equal(t, "", Reverse(""))
	assert.Equal(t, "retsbol", Reverse("lobster"))
}

func TestReverseHandler(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"/reverse?word=lobster",
		nil,
	)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(reverseHandler)
	handler.ServeHTTP(rr, req)

	status := rr.Code
	assert.Equal(t, http.StatusOK, status)

	expected := "retsbol"
	assert.Equal(t, expected, rr.Body.String())
}
