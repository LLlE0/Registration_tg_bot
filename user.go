package main

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Team     string `json:"team"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Step     int    `json:"step"`
	Username string `json:"username"`
	Time     string `json:"time"`
}

func (p *User) ToString() string {
	var banned int64
	envelope := "yes"
	Db.QueryRow("SELECT id FROM blocked_users WHERE id=?", fmt.Sprint(p.ID)).Scan(&banned)
	if banned == 0 {
		envelope = "no"
	}
	return "ID - " + fmt.Sprint(p.ID) + "\nName - " + p.Name + "\nTeam - " + p.Team + "\nPhone - " + p.Phone + "\nEmail - " + p.Email + "\nUsername - " + p.Username + "\nTime - " + p.Time + "\nIs banned? " + envelope + "\n"
}

func SaveUser(user User) error {
	query := `
    INSERT INTO users (id, name, team, phone, email, username, time)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	_, err := Db.Exec(query, user.ID, user.Name, user.Team, user.Phone, user.Email, user.Username, user.Time)
	return err
}

func DeleteUser(id string) (sql.Result, error) {
	res, err := Db.Exec("DELETE FROM users WHERE id=?", id)
	return res, err
}

func getUserByID(id int64) (*User, error) {
	query := `
    SELECT id, name, team, phone, email, username, time
    FROM users
    WHERE id=?
    `
	row := Db.QueryRow(query, id)

	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Team, &u.Phone, &u.Email, &u.Username, &u.Time)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
