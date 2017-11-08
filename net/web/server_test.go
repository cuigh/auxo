package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	const text = "OK"
	s := Default()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.AcquireContext(rec, req)
	ctx.Text(text)

	assert.Equal(t, http.StatusOK, rec.Code)
	bytes, err := ioutil.ReadAll(rec.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, text, string(bytes))
}
