package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mehdibo/godeploy/pkg/auth"
	"github.com/mehdibo/godeploy/pkg/db"
	"github.com/mehdibo/godeploy/pkg/env"
	"github.com/mehdibo/godeploy/pkg/messenger"
	"github.com/mehdibo/godeploy/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"io"
	"net/http/httptest"
	"testing"
)

var (
	adminUser = db.User{
		Username:    "admin",
		HashedToken: "admin",
		Role:        auth.RoleAdmin,
	}
)

type ServerTestSuite struct {
	suite.Suite
	server *Server
	tx     *gorm.DB
	dbConn *gorm.DB
	msn    *messenger.Messenger
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
	err = db.AutoMigrate(dbConn)
	if err != nil {
		s.T().Fatalf("Couldn't migrate database: %s", err.Error())
	}
	// Make sure db is clean
	tables := []string{
		"users",
		"http_tasks",
		"ssh_tasks",
		"tasks",
		"applications",
	}
	for _, table := range tables {
		dbConn.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE")
	}
	return dbConn
}

func (s *ServerTestSuite) getMessenger() *messenger.Messenger {
	// Load broker credentials
	brHost := env.Get("AMQP_HOST")
	brUser := env.Get("AMQP_USER")
	brPass := env.Get("AMQP_PASS")
	brPort := env.GetDefault("AMQP_PORT", "5672")
	if brHost == "" || brUser == "" || brPass == "" {
		s.T().Fatal("required broker credentials are not set, check your .env file")
	}
	msn, err := messenger.NewMessenger("amqp://" + brUser + ":" + brPass + "@" + brHost + ":" + brPort + "/")
	if err != nil {
		s.T().Fatal(err)
	}
	return msn
}

func loadFixtures(dbConn *gorm.DB) error {
	users := []db.User{
		{
			Username:    "admin",
			HashedToken: auth.HashToken("admin"),
			Role:        auth.RoleAdmin,
		},
	}
	applications := []db.Application{
		{
			Name:        "Test App 1",
			Description: "Some app to test with",
			Secret:      auth.HashToken("deploy_token"),
			Tasks: []db.Task{
				{
					Priority: 0,
					TaskType: db.TaskTypeHttp,
					HttpTask: &db.HttpTask{
						Method: "GET",
						Url:    "https://example.com",
					},
				},
				{
					Priority: 1,
					TaskType: db.TaskTypeSsh,
					SshTask: &db.SshTask{
						Username: "spoody",
						Host:     "localhost",
						Port:     22,
						Command:  "/update.sh",
					},
				},
			},
		},
		{
			Name:        "Test App 2",
			Description: "Some app to test with",
			Secret:      auth.HashToken("deploy_token"),
			Tasks: []db.Task{
				{
					Priority: 0,
					TaskType: db.TaskTypeSsh,
					SshTask: &db.SshTask{
						Username: "spoody",
						Host:     "localhost",
						Port:     22,
						Command:  "/update.sh",
					},
				},
			},
		},
	}
	for _, user := range users {
		res := dbConn.Create(&user)
		if res.Error != nil {
			return res.Error
		}
	}
	for _, app := range applications {
		res := dbConn.Create(&app)
		if res.Error != nil {
			return res.Error
		}
	}
	return nil
}

func (s *ServerTestSuite) SetupSuite() {
	s.dbConn = s.getDb()
	assert.NoError(s.T(), loadFixtures(s.dbConn))
}

func (s *ServerTestSuite) SetupTest() {
	s.tx = s.dbConn.Begin()
	s.msn = s.getMessenger()
	s.server = NewServer(s.tx, s.msn)
}

func (s *ServerTestSuite) TearDownTest() {
	err := s.msn.PurgeQueue(messenger.AppDeployQueue)
	if err != nil {
		s.T().Error(err)
	}
	err = s.msn.Close()
	if err != nil {
		s.T().Error(err)
	}
	s.tx.Rollback()
}

func prepareRequest(method string, uri string, body io.Reader, authUser *db.User) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	req := httptest.NewRequest(method, uri, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if authUser != nil {
		c.Set(auth.UserKey, *authUser)
	}
	return c, rec
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
