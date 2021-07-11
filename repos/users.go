package repos

import (
	"database/sql"
	"strings"
)

type User struct {
	ID       int64
	UserName string
	ChatID   int64
	Message  string
	Answer   []string
}

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) CollectData(user *User) error {
	//username string, chatid int64, message string, answer []string

	answ := strings.Join(user.Answer, ", ")
	data := `INSERT INTO users(username, chat_id, message, answer) VALUES($1, $2, $3, $4);`

	if _, err := u.DB.Exec(data, `@`+user.UserName, user.ChatID, user.Message, answ); err != nil {
		return err
	}

	return nil
}

func (u *UserModel) CreateTable() error {
	if _, err := u.DB.Exec(`CREATE TABLE users(ID SERIAL PRIMARY KEY, TIMESTAMP TIMESTAMP DEFAULT CURRENT_TIMESTAMP, USERNAME TEXT, CHAT_ID INT, MESSAGE TEXT, ANSWER TEXT);`); err != nil {
		return err
	}
	return nil
}

func (u *UserModel) GetNumberOfUsers() (int64, error) {

	var count int64

	row := u.DB.QueryRow("SELECT COUNT(DISTINCT username) FROM users;")
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
