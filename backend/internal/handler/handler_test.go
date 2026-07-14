package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/handler"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/service"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// newTestClient boots the real router stack and returns a request helper bound
// to one player whose points the test controls directly.
func newTestClient(t *testing.T, db *gorm.DB, points int) func(method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(middleware.EnsurePlayer(db, false))
	handler.New(service.New(db)).Register(api)

	player := testutil.NewPlayer(t, db, points)
	cookie := &http.Cookie{Name: "player_id", Value: player.ID.String()}

	return func(method, path, body string) *httptest.ResponseRecorder {
		var req *http.Request
		if body == "" {
			req = httptest.NewRequest(method, path, nil)
		} else {
			req = httptest.NewRequest(method, path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(cookie)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		return res
	}
}

func TestClaimEndpoint_StatusMapping(t *testing.T) {
	db := testutil.DB(t)
	do := newTestClient(t, db, 5000)

	cases := []struct {
		name string
		body string
		want int
	}{
		{"invalid body", `{"checkpoint": "abc"}`, http.StatusBadRequest},
		{"missing checkpoint", `{}`, http.StatusBadRequest},
		{"unknown checkpoint", `{"checkpoint": 6000}`, http.StatusBadRequest},
		{"not reached", `{"checkpoint": 7500}`, http.StatusConflict},
		{"success", `{"checkpoint": 5000}`, http.StatusOK},
		{"duplicate", `{"checkpoint": 5000}`, http.StatusConflict},
	}
	for _, tc := range cases {
		res := do(http.MethodPost, "/api/v1/claims", tc.body)
		if res.Code != tc.want {
			t.Errorf("%s: status = %d, want %d (body: %s)", tc.name, res.Code, tc.want, res.Body.String())
		}
	}
}

func TestResetEndpoint_ClearsState(t *testing.T) {
	db := testutil.DB(t)
	do := newTestClient(t, db, 10000)

	if res := do(http.MethodPost, "/api/v1/claims", `{"checkpoint": 10000}`); res.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want 200", res.Code)
	}
	if res := do(http.MethodPost, "/api/v1/reset", ""); res.Code != http.StatusNoContent {
		t.Fatalf("reset status = %d, want 204", res.Code)
	}

	res := do(http.MethodGet, "/api/v1/me", "")
	if res.Code != http.StatusOK {
		t.Fatalf("me status = %d, want 200", res.Code)
	}
	body := res.Body.String()
	if !strings.Contains(body, `"points":0`) || strings.Contains(body, `"claimed":true`) {
		t.Errorf("summary after reset = %s, want 0 points and nothing claimed", body)
	}

	hist := do(http.MethodGet, "/api/v1/history/claims", "")
	if !strings.Contains(hist.Body.String(), `"items":[]`) {
		t.Errorf("claim history after reset = %s, want empty items", hist.Body.String())
	}
}
