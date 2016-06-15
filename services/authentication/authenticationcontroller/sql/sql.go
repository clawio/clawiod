package sql

import (
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/services/authentication/authenticationcontroller"
	"github.com/clawio/clawiod/services/authentication/lib"
	_ "github.com/go-sql-driver/mysql" // enable mysql driver
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"           // enable postgresql driver
	_ "github.com/mattn/go-sqlite3" // enable sqlite3 driver
)

type controller struct {
	db            *gorm.DB
	authenticator *lib.Authenticator
	conf          *config.Config
}

// Options  holds the configuration
// parameters used by the controller.
type Options struct {
	Driver, DSN   string
	Authenticator *lib.Authenticator
}

// New returns an AuthenticationControler that uses a SQL database for handling
// users and JWT for tokens.
func New(conf *config.Config) (authenticationcontroller.AuthenticationController, error) {
	dirs := conf.GetDirectives()
	authenticator := lib.NewAuthenticator(dirs.Server.JWTSecret, dirs.Server.JWTSigningMethod)
	db, err := gorm.Open(dirs.Authentication.SQL.Driver, dirs.Authentication.SQL.DSN)
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&userRecord{}).Error
	if err != nil {
		return nil, err
	}

	return &controller{
		conf:          conf,
		db:            db,
		authenticator: authenticator,
	}, nil
}

func (c *controller) Authenticate(username, password string) (string, error) {
	rec, err := c.findByCredentials(username, password)
	if err != nil {
		return "", err
	}
	u := &entities.User{
		Username:    rec.Username,
		Email:       rec.Email,
		DisplayName: rec.DisplayName,
	}
	return c.authenticator.CreateToken(u)
}

// findByCredentials finds an user given an username and a password.
func (c *controller) findByCredentials(username, password string) (*userRecord, error) {
	rec := &userRecord{}
	err := c.db.Where("username=? AND password=?", username, password).First(rec).Error
	return rec, err
}

// TODO(labkode) set collation for table and column to utf8. The default is swedish
type userRecord struct {
	Username    string `gorm:"primary_key"`
	Email       string
	DisplayName string
	Password    string
}

func (u userRecord) TableName() string {
	return "users"
}
