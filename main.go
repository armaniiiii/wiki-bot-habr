package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	_ "github.com/lib/pq"
	"gitlab.com/armanbimak27/wiki-bot.git/repos"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	models "gitlab.com/armanbimak27/wiki-bot.git/models"
)

type AppWiki struct {
	models repos.Models
}

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("English", "en"),
		tgbotapi.NewInlineKeyboardButtonData("Russian", "ru"),
		tgbotapi.NewInlineKeyboardButtonData("Kazakh", "kk"),
	),
)

func telegramBot(app *AppWiki) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			fmt.Print(update)

			err := os.Setenv("LANGUAGE", update.CallbackQuery.Data)
			if err != nil {
				log.Println("can set lang")
			}
			//bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

			bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Language was set to "+update.CallbackQuery.Data))

		} else if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" && update.Message != nil {

			if update.Message.IsCommand() {
				if update.Message.Text == "/start" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a wikipedia bot, i can search information in a wikipedia, send me something what you want find in Wikipedia.")
					_, err2 := bot.Send(msg)
					if err2 != nil {
						return
					}
				} else if update.Message.Text == "/number_of_users" {

					num, err := app.models.Users.GetNumberOfUsers()
					if err != nil {

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error.")

						_, err := bot.Send(msg)
						if err != nil {
							log.Fatal(err.Error())
							return
						}
					}

					ans := fmt.Sprintf("%d peoples used me for search information in Wikipedia", num)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, ans)
					_, err = bot.Send(msg)
					if err != nil {
						log.Fatal(err.Error())
						return
					}

				} else if update.Message.Text == "/change_lang" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "please choose the lang")
					msg.ReplyMarkup = numericKeyboard
					_, err = bot.Send(msg)

					//if strings.Contains(update.Message.Text, "/change_lang") {
					//	lang := update.Message.CommandArguments()
					//	log.Println(lang)
					//}else{
					//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "command not supported")
					//	bot.Send(msg)
					//}
				}
			} else {

				language := os.Getenv("LANGUAGE")

				ms, _ := urlEncoded(update.Message.Text)

				urls := ms
				request := "https://" + language + ".wikipedia.org/w/api.php?action=opensearch&search=" + urls + "&limit=3&origin=*&format=json"

				message := wikipediaAPI(request)
				log.Println("array of request", message)

				///fuck

				user := repos.User{UserName: update.Message.Chat.UserName, ID: update.Message.Chat.ID, Message: update.Message.Text, Answer: message}

				if err := app.models.Users.CollectData(&user); err != nil {

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Database error, but bot still working.")
					bot.Send(msg)
				}

				for _, val := range message {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, val)
					bot.Send(msg)
				}
			}
		} else {

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Use the words for search.")
			bot.Send(msg)
		}
	}

}

func main() {
	time.Sleep(10 * time.Second)
	log.Println("waited db start")
	var host = os.Getenv("HOST")
	var port = os.Getenv("PORT")
	var user = os.Getenv("USER")
	var password = os.Getenv("PASSWORD")
	var dbname = os.Getenv("DBNAME")
	var sslmode = os.Getenv("SSLMODE")

	var dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dbInfo)

	if err != nil {
		log.Fatal(err.Error())
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}(db)

	app := AppWiki{
		models: repos.NewModels(db),
	}
	err = app.models.Users.CreateTable()
	log.Println("table created")

	if err != nil {
		log.Fatal(err.Error())

	}

	telegramBot(&app)

}

func wikipediaAPI(request string) (answer []string) {
	s := make([]string, 3)

	if response, err := http.Get(request); err != nil {
		log.Println("GET", err.Error())
		s[1] = "Wikipedia is not respond"
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)
		log.Println("GET Contents", string(contents))
		if err != nil {
			log.Fatal(err)
		}

		sr := &models.SearchResults{}
		if err = json.Unmarshal(contents, sr); err != nil {
			s[1] = "Something going wrong, try to change your question"
		}

		if !sr.Ready {
			s[1] = "Something going wrong, try to change your question"
		}

		for i := range sr.Results {
			s[i] = sr.Results[i].URL
		}
	}

	return s
}

func urlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
