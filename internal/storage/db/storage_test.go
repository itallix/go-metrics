package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DBStorageTestSuite struct {
	dbContainter testcontainers.Container

	suite.Suite
}

func (suite *DBStorageTestSuite) SetupSuite() {
	dbName := "metrics"
	dbUser := "username"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(context.Background(),
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	suite.Require().NoError(err)

	suite.dbContainter = postgresContainer
}

func (suite *DBStorageTestSuite) TearDownSuite() {
	suite.Require().NoError(suite.dbContainter.Terminate(context.Background()))
}

func (suite *DBStorageTestSuite) TestDbStorage() {
	ctx := context.Background()
	endpoint, err := suite.dbContainter.Endpoint(ctx, "")
	suite.Require().NoError(err)

	dsn := fmt.Sprintf("postgres://username:password@%s/metrics?sslmode=disable", endpoint)
	storage, err := NewPgStorage(ctx, dsn)
	suite.Require().NoError(err)
	suite.NotNil(storage)

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM   information_schema.tables 
			WHERE  table_schema = 'public'
			AND    table_name = $1
		);
	`
	err = storage.pool.QueryRow(context.Background(), query, "counters").Scan(&exists)
	suite.Require().NoError(err)
	suite.True(exists, "Table 'counters' was not created")

	err = storage.pool.QueryRow(context.Background(), query, "gauges").Scan(&exists)
	suite.Require().NoError(err)
	suite.True(exists, "Table 'gauges' was not created")
}

func TestDbStorageTestSuite(t *testing.T) {
	suite.Run(t, new(DBStorageTestSuite))
}
