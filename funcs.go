package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var Bot *tgbotapi.BotAPI
var AdminIDs []int64
var Db *sql.DB
var Step = 0
var UserData = make(map[int64]*User)
var IdList []string

// -----------------------Predicates----------------------//

// Predicate checks whether user is an admin
func IsAdmin(id int64) bool {
	for _, adminID := range AdminIDs {
		if id == adminID {
			return true
		}
	}
	return false
}

// Predicate checks whether user is banned
func IsBanned(message *tgbotapi.Message) bool {
	var id int64
	err := Db.QueryRow("SELECT id FROM blocked_users WHERE id=?", message.From.ID).Scan(&id)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Printf("Error checking if user is banned: %v", err)
		return false
	}
	return true
}

///////////////////////////////////////////////////////////

// ---------------------Handlers--------------------------//

/*
Function initiates the process of registration. It begins the dialogue with user which lasts for 4 messages.
Then it inserts user into the DB.
Feel free to change the dialogues here
*/
func StartRegistration(message *tgbotapi.Message) {
	if message.Text == "/start" {
		user, err := getUserByID(message.From.ID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error getting user by ID: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured while getting your data! We are working on it already!")
			Bot.Send(msg)
			return
		}

		if user != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "You are already registered")
			Bot.Send(msg)
			return
		}

		user = &User{ID: message.From.ID, Username: "@" + message.From.UserName, Time: time.Now().Format("2006-01-02 15:04:05")}
		msg := tgbotapi.NewMessage(message.Chat.ID, "Input your full name, please:")
		Bot.Send(msg)

		UserData[message.From.ID] = user

		return
	}
	if user, ok := UserData[message.From.ID]; ok {
		switch user.Step {
		case 0:
			user.Name = message.Text
			msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Nice to meet you, %s. Input your phone number:", user.Name))
			Bot.Send(msg)
			user.Step++
		case 1:
			user.Phone = message.Text
			msg := tgbotapi.NewMessage(message.Chat.ID, "What's your team name?")
			Bot.Send(msg)
			user.Step++
		case 2:
			user.Team = message.Text
			msg := tgbotapi.NewMessage(message.Chat.ID, "One message to go! Input your email, please: ")
			Bot.Send(msg)
			user.Step++
		case 3:
			user.Email = message.Text
			err := SaveUser(*user)
			if err != nil {
				log.Printf("Error saving user: %v", err)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured while inserting you into the database! Perhaps, you inserted some invalid data. Please, consider your data is correct or try again later. If it does not help, you may contact admins of the event.")
				Bot.Send(msg)
			} else {
				log.Print("User ", user.ID, " was registered!")
				msg := tgbotapi.NewMessage(message.Chat.ID, "Congrats! You had been successfully registered!")
				Bot.Send(msg)
			}
			delete(UserData, message.From.ID)
			log.Print("Registered user ", message.From.ID)
			return
		}
	}
}

/*
Function creates a txt file with the list of all registered users.
Then it sends it to the requested admin and deletes it.
*/
func SendList(message *tgbotapi.Message) {
	if !IsAdmin(message.From.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
		return
	}

	rows, err := Db.Query("SELECT id, name, team, phone, email, username, time FROM users ORDER BY time ASC")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured sending the list of participants! We are working on it!")
		Bot.Send(msg)
		return
	}
	defer rows.Close()

	var participants []string
	participants = append(participants, "~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

	for rows.Next() {
		var p User
		err := rows.Scan(&p.ID, &p.Name, &p.Team, &p.Phone, &p.Email, &p.Username, &p.Time)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		participants = append(participants, p.ToString())
	}

	if len(participants) == 1 {
		participants = append(participants, "Empty :(")
	}
	participants = append(participants, "~~~~~~~~~~~~~~~~~~~~~~~~~~~~")

	// Create a temporary file
	tempFile, err := os.Create("participants.txt")
	if err != nil {
		log.Printf("Error creating temporary file: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured creating the file with the list! We are working on it!")
		Bot.Send(msg)
		return
	}

	// Write the participants' information to the file
	_, err = tempFile.WriteString(strings.Join(participants, "\n"))
	if err != nil {
		log.Printf("Error writing to temporary file: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured creating the file with the list! We are working on it!")
		Bot.Send(msg)
		return
	}
	Z := tgbotapi.FilePath("participants.txt")
	// Send the file to the user
	fileSend := tgbotapi.RequestFile{
		Name: "participants.txt",
		Data: Z,
	}
	document := tgbotapi.NewDocument(message.Chat.ID, fileSend.Data)
	_, err = Bot.Send(document)
	if err != nil {
		log.Printf("Error sending document: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured sending the file with the list! We are working on it!")
		Bot.Send(msg)
		return
	}

	tempFile.Close()
	err = os.Remove("participants.txt")
	if err != nil {
		log.Printf("Failed to remove temp file: %s", err)
	}

	log.Print("Sent list")
}

/*
Function deletes the user from the DB using its ID.
It parses the command and deletes the row from the table of users.
*/
func UserDelete(message *tgbotapi.Message) {
	if !IsAdmin(message.From.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
		return
	}
	id := message.Text[5:]

	res, err := DeleteUser(id)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured while deleting user! Consider sending valid ID, please.")
		Bot.Send(msg)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error getting deleted rows count: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured while deleting user! We are working on it!")
		Bot.Send(msg)
		return
	}

	if count == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "No user with such ID.")
		Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "The user had been deleted successfully")
	log.Print("Deleted user ", id)
	Bot.Send(msg)
}

/*
Function makes bot reusable.
Deletes the users and banned users table.
*/
func DropTable(message *tgbotapi.Message) {
	if !IsAdmin(message.From.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
		return
	}

	_, err := Db.Exec("DELETE FROM users")
	if err != nil {
		log.Printf("Error dropping db: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error deleting the list of users")
		Bot.Send(msg)
		return
	}

	_, err = Db.Exec("DELETE FROM blocked_users")
	if err != nil {
		log.Printf("Error dropping banned db: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error deleting the list of banned users")
		Bot.Send(msg)
		return
	}

	log.Print("Dropped DB")
	msg := tgbotapi.NewMessage(message.Chat.ID, "The database had been cleaned")
	Bot.Send(msg)
}

/*
Function sends messages to the users with provided IDs.
It recieves the message in format:
/send | [all | {list_of_ids}] | template
It also changes strings {name}, {time}, {team}, {email} to the values of the reciever

Example:
/send | all | Hi, {name}
/send | 1234567890 1111111111 | Hi, user! You registered at {time} with email {email} and your team name is {team}
*/
func Dispatch(message *tgbotapi.Message) {
	if !IsAdmin(message.From.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
		return
	}

	parts := strings.Split(message.Text, "|")
	IdList := strings.Split(strings.TrimSpace(parts[1]), " ")
	template := strings.TrimSpace(parts[2])
	if IdList[0] == "all" {
		rows, err := Db.Query("SELECT id, name, team, email, time FROM users")
		if err != nil {
			log.Printf("Error querying users: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Error sending messages, we are working on it!")
			Bot.Send(msg)
			return
		}
		defer rows.Close()
		var userFromDB User
		for rows.Next() {
			err := rows.Scan(&userFromDB.ID, &userFromDB.Name, &userFromDB.Team, &userFromDB.Email, &userFromDB.Time)
			if err != nil {
				log.Printf("Error scanning user: %v", err)
				continue
			}

			msgText := template
			msgText = strings.Replace(msgText, "{name}", userFromDB.Name, -1)
			msgText = strings.Replace(msgText, "{time}", userFromDB.Time, -1)
			msgText = strings.Replace(msgText, "{team}", userFromDB.Team, -1)
			msgText = strings.Replace(msgText, "{email}", userFromDB.Email, -1)

			msg := tgbotapi.NewMessage(userFromDB.ID, msgText)
			_, err = Bot.Send(msg)
			if err != nil {
				log.Print("Error sending message: ", err)
			}

			log.Print("Sending message to: ", userFromDB.ID)
			time.Sleep(2 * time.Second)
		}
	} else {
		for _, idStr := range IdList {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				log.Printf("Error parsing user ID: %v", err)
				continue
			}

			userFromDB, err := getUserByID(id)
			if err != nil {
				log.Printf("Error getting user by ID: %v", err)
				continue
			}

			msgText := template
			msgText = strings.Replace(msgText, "{name}", userFromDB.Name, -1)
			msgText = strings.Replace(msgText, "{time}", userFromDB.Time, -1)
			msgText = strings.Replace(msgText, "{team}", userFromDB.Team, -1)
			msgText = strings.Replace(msgText, "{email}", userFromDB.Email, -1)

			msg := tgbotapi.NewMessage(id, msgText)
			_, err = Bot.Send(msg)
			if err != nil {
				log.Print("Error sending message: ", err)
			}

			log.Print("Sending message to: ", id)
			time.Sleep(2 * time.Second)
		}
	}
}

/*
Function sends the amount of participants
*/
func CountPeople(message *tgbotapi.Message) {
	if !IsAdmin(message.From.ID) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
		return
	}
	var count int
	rows, err := Db.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		log.Printf("Error querying users count: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error counting participants! We are working on it!")
		Bot.Send(msg)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Printf("Error scanning users count: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Error counting participants! We are working on it!")
			Bot.Send(msg)
			return
		}
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Total members: %d", count))
	log.Print("Get count")
	Bot.Send(msg)
}

/*
Function adds the provided user ID to the blocked_users table
*/
func Block(message *tgbotapi.Message) {
	if IsAdmin(message.From.ID) && len(message.Text) > 4 && message.Text[:4] == "/ban" {
		userID := message.Text[5:]

		_, err := Db.Exec("INSERT INTO blocked_users (id) VALUES (?)", userID)
		if err != nil {
			log.Printf("Error blocking user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured while banning user! Perhaps, they were banned already.")
			Bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "The user had been blocked.")
		Bot.Send(msg)
		log.Print("Banned user ", userID)
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
	}
}

/*
Function deletes the provided user ID from the blocked_users table
*/
func UnBlock(message *tgbotapi.Message) {
	if IsAdmin(message.From.ID) && len(message.Text) > 6 && message.Text[:6] == "/unban" {
		userID := message.Text[7:]

		_, err := Db.Exec("DELETE FROM blocked_users WHERE id = ?", userID)
		if err != nil {
			log.Printf("Error blocking user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Error occured while unbanning user! Perhaps, they were unbanned already.")
			Bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "The user had been unblocked.")
		Bot.Send(msg)
		log.Print("Banned user ", userID)
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown menu choice! Type /help to see the allowed commands for you!")
		log.Print("Access denied")
		Bot.Send(msg)
	}
}

/*
Function sends the help-message depending on whether the user is an admin of the bot
*/
func SendHelpMessage(message *tgbotapi.Message) {
	var msgText string
	if !IsAdmin(message.From.ID) {
		msgText = "Bot created by https://github.com/LLlE0\nWrite /start to register or check your registration status!"
	} else {
		msgText = `Bot created by https://github.com/LLlE0

		/start - registration
		/count - get the amount of registered users
		/list - get the list of registered users
		/del_{id} - delete the person with the id from the list
		/deleteall - clear the database, don't forget to save it as list!
		/ban_{id} - forbid a user to use your bot ever again

		/send | {id1} {id2} ... | Hi, {name}! - send the message to user with specific ids. Here, {name} stands for the name user stated in his application form. The fields to change are [{name}, {team}, {phone}, {email}]
		
		/send | all | Hi, {name}! - send the message to all users
		`
	}
	msg := tgbotapi.NewMessage(message.From.ID, msgText)
	Bot.Send(msg)
	log.Print("Help message sent")
}
