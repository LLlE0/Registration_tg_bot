package main

import (
	"bufio"
	"database/sql"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open(".env")
	if err != nil {
		log.Print("Error reading .env file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Print("Error reading .env file:", err)
		return
	}

	botToken := os.Getenv("BOT_TOKEN")
	Bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", Bot.Self.UserName)

	Db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Database service is up")
	defer Db.Close()

	query, err := os.ReadFile("db.sql")
	if err != nil {
		log.Fatal("Error creating DB: ", err)
	}
	_, err = Db.Exec(string(query))
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
		AdminIDs = append(AdminIDs, id)
	}
	log.Print("Moderators IDs: ", AdminIDs)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go handleMessage(update.Message)
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	log.Printf("%d -> sent %s\n", message.From.ID, message.Text)
	if IsBanned(message) {
		log.Print("But they were banned")
		msg := tgbotapi.NewMessage(message.Chat.ID, "You were banned in this bot!")
		Bot.Send(msg)
		return
	}

	_, ok := UserData[message.From.ID]
	if message.Text == "/start" || ok {
		StartRegistration(message)
	} else if message.Text == "/list" {
		SendList(message)
	} else if message.Text == "/deleteall" {
		DropTable(message)
	} else if strings.HasPrefix(message.Text, "/del") {
		UserDelete(message)
	} else if strings.HasPrefix(message.Text, "/send") {
		Dispatch(message)
	} else if message.Text == "/count" {
		CountPeople(message)
	} else if message.Text == "/help" {
		SendHelpMessage(message)
	} else if strings.HasPrefix(message.Text, "/ban") {
		Block(message)
	} else if strings.HasPrefix(message.Text, "/unban") {
		UnBlock(message)
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		Bot.Send(msg)
	}
}
