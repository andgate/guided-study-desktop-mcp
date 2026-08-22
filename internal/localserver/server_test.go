package localserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutesHealthProbe(t *testing.T) {
	server := New("127.0.0.1:7331", http.NotFoundHandler())
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ok\n", response.Body.String())

	contentType := response.Header().Get("Content-Type")
	require.Equal(t, "text/plain; charset=utf-8", contentType)
}
