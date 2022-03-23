package main

import (
	"context"
	"database/sql"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type mysqlSessionRepository struct {
	mysql *sql.DB
}

func NewSessionRepository(mysql *sql.DB) SessionRepository {
	return mysqlSessionRepository{
		mysql: mysql,
	}
}

func (m mysqlSessionRepository) FindUserIdByUuid(ctx context.Context, uuid string) (int64, error) {
	stmt, err := m.mysql.Prepare("SELECT user FROM sessions WHERE uuid = ?")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var userId int64
	ctx = newrelic.NewContext(ctx, newrelic.FromContext(ctx))
	err = stmt.QueryRowContext(ctx, uuid).Scan(&userId)

	if err != nil {
		return -1, err
	}
	return userId, nil
}

func (m mysqlSessionRepository) UpdateLastUsedByUuid(uuid string) error {
	stmt, err := m.mysql.Prepare("UPDATE sessions SET last_used = now() WHERE uuid = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(uuid)
	if err != nil {
		return err
	}
	return nil
}
