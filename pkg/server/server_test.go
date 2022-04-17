package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/go_deploy/pkg/auth"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/mehdibo/go_deploy/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
	Server *Server
}

func (s *ServerTestSuite) getDb() *gorm.DB {
	// Load database credentials
	dbHost := env.Get("DB_HOST")
	dbUser := env.Get("DB_USER")
	dbPass := env.Get("DB_PASS")
	dbName := env.Get("DB_NAME")
	dbPort := env.GetDefault("DB_PORT", "5432")
	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" {
		s.T().Fatal("Required database credentials are not set")
	}
	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort
	dbConn, err := db.NewDb(dsn)
	if err != nil {
		s.T().Fatalf("Couldn't connect to test database: %s", err.Error())
	}
	// Make sure db is clean
	tables := []string{
		"users",
	}
	for _, table := range tables {
		dbConn.Exec("TRUNCATE " + table + " RESTART IDENTITY")
	}
	return dbConn
}

func loadFixtures(dbConn *gorm.DB) {
	users := []db.User{
		{
			Username:    "admin",
			HashedToken: auth.HashToken("admin"),
			Role:        auth.RoleAdmin,
		},
	}
	for _, user := range users {
		dbConn.Create(&user)
	}
}

func (s *ServerTestSuite) SetupTest() {
	dbConn := s.getDb()
	loadFixtures(dbConn)
	s.Server = NewServer(dbConn)

}

func prepareRequest(method string, uri string, body io.Reader, authUser *db.User) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, uri, body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if authUser != nil {
		c.Set(auth.UserKey, *authUser)
	}
	return c, rec
}

func (s *ServerTestSuite) TestPing() {
	s.T().Run("unauthenticated", func(t *testing.T) {
		ctx, rec := prepareRequest(http.MethodGet, "/api/ping", nil, nil)
		if assert.NoError(t, s.Server.Ping(ctx)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
		}

	})
	s.T().Run("authenticated admin", func(t *testing.T) {
		adminUser := db.User{
			Username:    "admin",
			HashedToken: "admin",
			Role:        auth.RoleAdmin,
		}
		ctx, rec := prepareRequest(http.MethodGet, "/api/ping", nil, &adminUser)
		if assert.NoError(t, s.Server.Ping(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "{\"message\":\"pong\"}\n", rec.Body.String())
		}

	})
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
