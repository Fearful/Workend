package testutil

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"workend/api/internal/db"
)

var (
	sharedPool *pgxpool.Pool
	sharedOnce sync.Once
	sharedErr  error
)

func SetupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	sharedOnce.Do(func() {
		ctx := context.Background()

		pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
			postgres.WithDatabase("workend_test"),
			postgres.WithUsername("test"),
			postgres.WithPassword("test"),
			testcontainers.WithWaitStrategy(
				wait.ForAll(
					wait.ForLog("database system is ready to accept connections").
						WithOccurrence(2).
						WithStartupTimeout(60*time.Second),
					wait.ForListeningPort("5432/tcp").
						WithStartupTimeout(60*time.Second),
				),
			),
		)
		if err != nil {
			sharedErr = fmt.Errorf("start postgres container: %w", err)
			return
		}

		connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			sharedErr = fmt.Errorf("get connection string: %w", err)
			return
		}

		if err := db.Migrate(connStr); err != nil {
			sharedErr = fmt.Errorf("run migrations: %w", err)
			return
		}

		pool, err := db.Connect(ctx, connStr)
		if err != nil {
			sharedErr = fmt.Errorf("connect pool: %w", err)
			return
		}

		sharedPool = pool
	})

	if sharedErr != nil {
		t.Fatalf("testdb setup: %v", sharedErr)
	}
	return sharedPool
}

func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		DO $$
		DECLARE r RECORD;
		BEGIN
			FOR r IN (
				SELECT tablename FROM pg_tables
				WHERE schemaname = 'public' AND tablename != 'goose_db_version'
			) LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$
	`)
	if err != nil {
		t.Fatalf("truncate all tables: %v", err)
	}
}
