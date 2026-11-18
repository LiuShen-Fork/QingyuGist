package api_test

import (
	"testing"

	webtest "github.com/LiuShen-Fork/QingyuGist/internal/web/test"
	"github.com/stretchr/testify/require"
)

func TestOpenAPISpec(t *testing.T) {
	s := webtest.Setup(t)
	defer webtest.Teardown(t)

	resp := s.Request(t, "GET", "/api/openapi.yaml", nil, 200)
	require.Contains(t, resp.Header.Get("Content-Type"), "yaml")
}
