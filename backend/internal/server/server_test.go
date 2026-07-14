package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/config"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/server"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/testutil"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func testApp(db *gorm.DB) *fiber.App {
	return server.New(db, config.Config{
		FrontendOrigin: "http://localhost:3000",
		CookieSecure:   false,
	})
}

func playerCookie(t *testing.T, res *http.Response) *http.Cookie {
	t.Helper()
	for _, ck := range res.Cookies() {
		if ck.Name == "player_id" {
			return ck
		}
	}
	return nil
}

func cleanupPlayer(t *testing.T, db *gorm.DB, id string) {
	t.Cleanup(func() {
		db.Delete(&model.Player{}, "id = ?", id)
	})
}

func TestEnsurePlayer_BootstrapsNewPlayer(t *testing.T) {
	db := testutil.DB(t)
	app := testApp(db)

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/me", nil), -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	ck := playerCookie(t, res)
	if ck == nil {
		t.Fatal("expected player_id cookie to be set")
	}
	if !ck.HttpOnly {
		t.Error("cookie must be HttpOnly")
	}
	cleanupPlayer(t, db, ck.Value)

	var count int64
	db.Model(&model.Player{}).Where("id = ?", ck.Value).Count(&count)
	if count != 1 {
		t.Fatalf("player row count = %d, want 1", count)
	}
}

func TestEnsurePlayer_ReusesExistingPlayer(t *testing.T) {
	db := testutil.DB(t)
	app := testApp(db)

	first, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/me", nil), -1)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	ck := playerCookie(t, first)
	if ck == nil {
		t.Fatal("expected player_id cookie on first request")
	}
	cleanupPlayer(t, db, ck.Value)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: ck.Value})
	second, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}

	if second.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", second.StatusCode)
	}
	if playerCookie(t, second) != nil {
		t.Error("existing player should not get a new cookie")
	}
}

func TestEnsurePlayer_TamperedCookieGetsFreshPlayer(t *testing.T) {
	db := testutil.DB(t)
	app := testApp(db)

	fake := uuid.NewString() // valid UUID but no matching player row
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: fake})
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	ck := playerCookie(t, res)
	if ck == nil {
		t.Fatal("expected a fresh player cookie")
	}
	cleanupPlayer(t, db, ck.Value)
	if ck.Value == fake {
		t.Error("must not adopt the tampered player ID")
	}
}
