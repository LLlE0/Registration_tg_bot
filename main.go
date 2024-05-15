package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Team  string `json:"team"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Step  int    `json:"step"`
}

func (p *User) ToString() string {
	return "ID-" + fmt.Sprint(p.ID) + ": \nФИО - " + p.Name + "\nКоманда - " + p.Team + "\nТелефон - " + p.Phone + "\n Почта - " + p.Email + "\n"
}

var bot *tgbotapi.BotAPI
var adminIDs []int64
var db *sql.DB
var userData = make(map[int64]*User)

func saveUser(user User) error {
	query := `
    INSERT OR IGNORE INTO users (id, name, team, phone, email)
    VALUES (?, ?, ?, ?, ?)
    `
	_, err := db.Exec(query, user.ID, user.Name, user.Team, user.Phone, user.Email)
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("BOT_TOKEN")
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTableQuery, err := os.ReadFile("db.sql")
	if err != nil {
		log.Fatal("Error creating DB: ", err)
	}
	_, err = db.Exec(string(createTableQuery))
	if err != nil {
		log.Fatal(err)
	}
	adminIDStr := os.Getenv("ADMIN_IDS")

	for _, idStr := range strings.Split(adminIDStr, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing admin ID: %v", err)
			continue
		}
		adminIDs = append(adminIDs, id)
	}
	log.Print("Moderators IDs: ", adminIDs)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go handleMessage(update.Message)
		}
	}
}

func getUserByID(id int64) (*User, error) {
	query := `
    SELECT id, name, team, phone, email
    FROM users
    WHERE id=?
    `
	row := db.QueryRow(query, id)

	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Team, &u.Phone, &u.Email)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func handleMessage(message *tgbotapi.Message) {
	if message.Text == "/start" {
		user, err := getUserByID(message.From.ID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error getting user by ID: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Возникла ошибка при загрузке ваших данных, мы уже работаем над вопросом!")
			bot.Send(msg)
			return
		}

		if user != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Вы уже зарегистрированы! Список всех участников - /list")
			bot.Send(msg)
			return
		}

		user = &User{ID: message.From.ID}
		msg := tgbotapi.NewMessage(message.Chat.ID, "Введите ваше ФИО")
		bot.Send(msg)

		userData[message.From.ID] = user

		return
	}
	if user, ok := userData[message.From.ID]; ok {
		switch user.Step {
		case 0:
			user.Name = message.Text
			msg := tgbotapi.NewMessage(message.Chat.ID, "Введите свой контактный номер телефона")
			bot.Send(msg)
			user.Step++
		case 1:
			user.Phone = message.Text
			msg := tgbotapi.NewMessage(message.Chat.ID, "Введите название вашей команды")
			bot.Send(msg)
			user.Step++
		case 2:
			user.Team = message.Text
			msg := tgbotapi.NewMessage(message.Chat.ID, "Введите ваш контактный адрес эл. почты")
			bot.Send(msg)
			user.Step++
		case 3:
			user.Email = message.Text
			err := saveUser(*user)
			if err != nil {
				log.Printf("Error saving user: %v", err)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Возникла ошибка при занесении вас в базу данных, мы уже работаем над вопросом!")
				bot.Send(msg)
			} else {
				log.Print("User ", user.ID, " was registered!")
				msg := tgbotapi.NewMessage(message.Chat.ID, "Вы успешно зарегистрированы! Список всех участников можно увидеть с помощью команды /list")
				bot.Send(msg)
			}
			delete(userData, message.From.ID)
			return
		}
	} else if message.Text == "/list" {
		rows, err := db.Query("SELECT id, name, team, phone, email FROM users")
		if err != nil {
			log.Printf("Error querying users: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Возникла ошибка при загрузке списка участников, мы уже работаем над вопросом!")
			bot.Send(msg)
			return
		}
		defer rows.Close()

		var participants []string
		participants = append(participants, "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

		for rows.Next() {
			var p User
			err := rows.Scan(&p.ID, &p.Name, &p.Team, &p.Phone, &p.Email)
			if err != nil {
				log.Printf("Error scanning user: %v", err)
				continue
			}
			participants = append(participants, p.ToString())
		}

		if len(participants) == 1 {
			participants = append(participants, "Пока-что здесь пусто")
		}
		participants = append(participants, "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

		msg := tgbotapi.NewMessage(message.Chat.ID, "Список участников:")
		msg.Text += "\n" + strings.Join(participants, "\n")
		bot.Send(msg)
		return
	} else if len(message.Text) > 5 && message.Text[:5] == "/del_" {
		id, err := strconv.ParseInt(message.Text[5:], 10, 64)
		if err != nil {
			log.Printf("Error parsing user ID: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка при удалении участника")
			bot.Send(msg)
			return
		}

		if !isAdmin(message.From.ID) {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Недостаточно прав для удаления участника")
			bot.Send(msg)
			return
		}

		res, err := db.Exec("DELETE FROM users WHERE id=?", id)
		if err != nil {
			log.Printf("Error deleting user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка при удалении участника")
			bot.Send(msg)
			return
		}

		count, err := res.RowsAffected()
		if err != nil {
			log.Printf("Error getting deleted rows count: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка при удалении участника")
			bot.Send(msg)
			return
		}

		if count == 0 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Участник с таким ID не найден")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "Участник успешно удален")
		bot.Send(msg)
		return
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда, допустимые команды: /list и /start")
		bot.Send(msg)
	}
}

func isAdmin(id int64) bool {
	for _, adminID := range adminIDs {
		if id == adminID {
			return true
		}
	}
	return false
}
