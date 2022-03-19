package runelite

import "database/sql"

type mysqlSessionRepository struct {
	mysql *sql.DB
}

func NewSessionRepository(mysql *sql.DB) SessionRepository {
	return mysqlSessionRepository{
		mysql: mysql,
	}
}

func (m mysqlSessionRepository) FindUserIdByUuid(uuid string) (int64, error) {
	stmt, err := m.mysql.Prepare("SELECT user FROM sessions WHERE uuid = ?")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var userId int64
	err = stmt.QueryRow(uuid).Scan(&userId)

	if err != nil {
		return -1, err
	}
	return userId, nil
}
