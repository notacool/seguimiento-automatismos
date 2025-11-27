package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// SetupTestGin configura Gin en modo test.
func SetupTestGin() {
	gin.SetMode(gin.TestMode)
}

// MakeRequest realiza una petición HTTP de prueba.
func MakeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// ParseJSONResponse decodifica la respuesta JSON.
func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	t.Helper()

	err := json.Unmarshal(w.Body.Bytes(), target)
	require.NoError(t, err)
}

// AssertJSONResponse verifica el código de estado y decodifica la respuesta.
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	t.Helper()

	require.Equal(t, expectedStatus, w.Code)
	ParseJSONResponse(t, w, target)
}
