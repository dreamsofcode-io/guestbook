package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dreamsofcode-io/guestbook/internal/config"
)

func TestCreatingNewValidation(t *testing.T) {
	envs := []string{
		"PGUSER", "PGPASSWORD", "PGPORT", "PGDATABASE", "PGHOST",
	}

	fullSetup := func() {
		os.Setenv("PGUSER", "user")
		os.Setenv("PGPASSWORD", "password")
		os.Setenv("PGPORT", "5432")
		os.Setenv("PGDATABASE", "database")
		os.Setenv("PGHOST", "database.com")
	}

	clear := func() {
		for _, env := range envs {
			os.Unsetenv(env)
		}
	}

	testCases := []struct {
		Description string
		Setup       func()
		ExpectedCfg *config.Database
		ExpectedErr error
	}{
		{
			Description: "testing a complete env setup",
			Setup: func() {
				fullSetup()
			},
			ExpectedCfg: &config.Database{
				Username: "user",
				Password: "password",
				Port:     5432,
				Host:     "database.com",
				DBName:   "database",
			},
		},
		{
			Description: "no setup",
			Setup: func() {
				clear()
			},
			ExpectedErr: fmt.Errorf("no PGUSER env variable set"),
		},
		{
			Description: "no pg user",
			Setup: func() {
				fullSetup()
				os.Unsetenv("PGUSER")
			},
			ExpectedErr: fmt.Errorf("no PGUSER env variable set"),
		},
		{
			Description: "no pg password",
			Setup: func() {
				fullSetup()
				os.Unsetenv("PGPASSWORD")
			},
			ExpectedErr: fmt.Errorf("no PGPASSWORD env variable set"),
		},
		{
			Description: "no pg host",
			Setup: func() {
				fullSetup()
				os.Unsetenv("PGHOST")
			},
			ExpectedErr: fmt.Errorf("no PGHOST env variable set"),
		},
		{
			Description: "no pg port",
			Setup: func() {
				fullSetup()
				os.Unsetenv("PGPORT")
			},
			ExpectedErr: fmt.Errorf("no PGPORT env variable set"),
		},
		{
			Description: "no pg database",
			Setup: func() {
				fullSetup()
				os.Unsetenv("PGDATABASE")
			},
			ExpectedErr: fmt.Errorf("no PGDATABASE env variable set"),
		},
		{
			Description: "invalid port setup",
			Setup: func() {
				fullSetup()
				os.Setenv("PGPORT", "helloworld")
			},
			ExpectedErr: fmt.Errorf(
				"failed to convert port to int: strconv.Atoi: parsing \"helloworld\": invalid syntax",
			),
		},
		{
			Description: "empty db name",
			Setup: func() {
				fullSetup()
				os.Setenv("PGDATABASE", "")
			},
			ExpectedErr: fmt.Errorf(
				"failed to validate config: invalid database name",
			),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.Description, func(t *testing.T) {
			test.Setup()

			cfg, err := config.NewDatabase()

			if test.ExpectedErr != nil {
				assert.EqualError(t, err, test.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.ExpectedCfg, cfg)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	valid := config.Database{
		Username: "pguser",
		Password: "pgpassword",
		Host:     "pghost",
		Port:     5432,
		DBName:   "pgdatabase",
	}

	t.Run("Test valid", func(t *testing.T) {
		valid := valid
		assert.NoError(t, valid.Validate())
	})

	t.Run("Test invalid username", func(t *testing.T) {
		valid := valid
		valid.Username = ""
		assert.EqualError(t, valid.Validate(), "invalid username")
	})

	t.Run("Test invalid password", func(t *testing.T) {
		valid := valid
		valid.Password = ""
		assert.EqualError(t, valid.Validate(), "invalid password")
	})

	t.Run("Test invalid host", func(t *testing.T) {
		valid := valid
		valid.Host = ""
		assert.EqualError(t, valid.Validate(), "invalid host")
	})

	t.Run("Test invalid port", func(t *testing.T) {
		valid := valid
		valid.Port = 0
		assert.EqualError(t, valid.Validate(), "invalid port")
	})

	t.Run("Test invalid name", func(t *testing.T) {
		valid := valid
		valid.DBName = ""
		assert.EqualError(t, valid.Validate(), "invalid database name")
	})
}
