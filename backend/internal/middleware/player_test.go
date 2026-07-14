package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Eursukkul/Points-Collection-Game/backend/internal/database"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/middleware"
	"github.com/Eursukkul/Points-Collection-Game/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://points:points_local@localhost:5432/points_game?sslmode=disable"
	}
	db, err := database.Connect(url)
	if err != nil {
		t.Skipf("postgres not available (run `docker compose up -d`): %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/", middleware.EnsurePlayer(db, false))
	api.GET("/whoami", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"playerId": middleware.PlayerID(c).String()})
	})
	return r
}

func playerCookie(t *testing.T, res *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, ck := range res.Result().Cookies() {
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
	db := testDB(t)
	r := testRouter(db)

	res := httptest.NewRecorder()
	r.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/whoami", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
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
	db := testDB(t)
	r := testRouter(db)

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/whoami", nil))
	ck := playerCookie(t, first)
	if ck == nil {
		t.Fatal("expected player_id cookie on first request")
	}
	cleanupPlayer(t, db, ck.Value)

	second := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: ck.Value})
	r.ServeHTTP(second, req)

	if second.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", second.Code)
	}
	if playerCookie(t, second) != nil {
		t.Error("existing player should not get a new cookie")
	}
}

func TestEnsurePlayer_TamperedCookieGetsFreshPlayer(t *testing.T) {
	db := testDB(t)
	r := testRouter(db)

	fake := uuid.NewString() // valid UUID but no matching player row
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: fake})
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
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
