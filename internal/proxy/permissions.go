package proxy

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

const (
	tokenExpirationDays = 30
)

const (
	getAuthUserSql = `
		SELECT user_id
		FROM users_token tokens 
		WHERE tokens.key = $1 AND tokens.created > $2
	`
	checkProjectAccessSql = `
		SELECT count(*)
		FROM projects_project projects
		WHERE projects.db_name = $1
	`
)

func isAccessAllowed(pgPool *pgxpool.Pool, database string, auth string) (allowed bool, err error) {
	if database == "" {
		return true, nil
	}

	conn, err := pgPool.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()

	if auth != "" {
		allowed, err = checkAccess(conn, database, auth)
	}

	return
}

func checkAccess(conn *pgxpool.Conn, database string, auth string) (allowed bool, err error) {
	user, err := retrieveUser(conn, auth)
	if err != nil {
		return
	}

	if user == -1 {
		return
	}

	var rowsCount *int
	err = conn.QueryRow(context.Background(), checkProjectAccessSql, database).Scan(&rowsCount)
	if err != nil {
		return
	}
	return *rowsCount > 0, nil
}

func retrieveUser(conn *pgxpool.Conn, auth string) (user int, err error) {
	now := time.Now()
	expiration := now.AddDate(0, 0, -tokenExpirationDays)

	var userId *int
	err = conn.QueryRow(context.Background(), getAuthUserSql, auth, expiration).Scan(&userId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, nil
		}
		return
	}
	return *userId, nil
}
