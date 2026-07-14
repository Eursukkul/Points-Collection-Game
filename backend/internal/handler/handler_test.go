package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/config"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/server"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
	"gorm.io/gorm"
)

// newTestClient boots the real app stack and returns a request helper bound
// to one player whose points the test controls directly.
func newTestClient(t *testing.T, db *gorm.DB, points int) func(method, path, body string) *http.Response {
	t.Helper()
	app := server.New(db, config.Config{
		FrontendOrigin: "http://localhost:3000",
		CookieSecure:   false,
	})

	player := testutil.NewPlayer(t, db, points)
	cookie := &http.Cookie{Name: "player_id", Value: player.ID.String()}

	return func(method, path, body string) *http.Response {
		var req *http.Request
		if body == "" {
			req = httptest.NewRequest(method, path, nil)
		} else {
			req = httptest.NewRequest(method, path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
		}
		req.AddCookie(cookie)
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		return res
	}
}

func readBody(t *testing.T, res *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
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
		if res.StatusCode != tc.want {
			t.Errorf("%s: status = %d, want %d (body: %s)", tc.name, res.StatusCode, tc.want, readBody(t, res))
		}
	}
}

func TestResetEndpoint_ClearsState(t *testing.T) {
	db := testutil.DB(t)
	do := newTestClient(t, db, 10000)

	if res := do(http.MethodPost, "/api/v1/claims", `{"checkpoint": 10000}`); res.StatusCode != http.StatusOK {
		t.Fatalf("claim status = %d, want 200", res.StatusCode)
	}
	if res := do(http.MethodPost, "/api/v1/reset", ""); res.StatusCode != http.StatusNoContent {
		t.Fatalf("reset status = %d, want 204", res.StatusCode)
	}

	res := do(http.MethodGet, "/api/v1/me", "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("me status = %d, want 200", res.StatusCode)
	}
	body := readBody(t, res)
	if !strings.Contains(body, `"points":0`) || strings.Contains(body, `"claimed":true`) {
		t.Errorf("summary after reset = %s, want 0 points and nothing claimed", body)
	}

	hist := do(http.MethodGet, "/api/v1/history/claims", "")
	if got := readBody(t, hist); !strings.Contains(got, `"items":[]`) {
		t.Errorf("claim history after reset = %s, want empty items", got)
	}
}
