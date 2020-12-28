package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"./botconfd"
	"github.com/fatih/structs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mitchellh/mapstructure"
)

var (
	users     map[string]interface{} // map ник - статус
	chatID    int64
	names     map[string]interface{} // map имя - ник
	listUsers map[string]interface{} // map имя - статус, для отчета
	statusMsg string
)

type nameFromJson struct { //под этими именами сохраняется в структуре. Имена переменных должны совпадать с key map names
	Руслан string
	//	Катя   string
	//	Молли  string
}

func main() {

	//fileJson = "state.json"

	//Ruslan := botconfd.GRR{}
	users = make(map[string]interface{}) // ники в telegram
	//users[botconfd.SDM] = "def_home"
	users[botconfd.GRR] = "def_home"
	//users[botconfd.AAA] = "def_home"
	//users[botconfd.SAY] = "def_home"
	//users[botconfd.DDP] = "def_home"
	//users[botconfd.AIY] = "def_home"

	names = make(map[string]interface{}) // настоящие имена
	//names["Дима"] = botconfd.SDM
	names["Руслан"] = botconfd.GRR
	//names["Айтуган"] = botconfd.AAA
	//names["Айдар"] = botconfd.SAY
	//names["Денис"] = botconfd.DDP
	//names["Ильнур"] = botconfd.AIY

	listUsers = make(map[string]interface{}) // map для отчета
	listUsers = readFromFile(botconfd.FileJson)

	// подключаемся к боту с помощью токена
	bot, err := tgbotapi.NewBotAPI(botconfd.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go sendReport(bot)

	// инициализируем канал, куда будут прилетать обновления от API
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	// читаем обновления из канала
	for update := range updates {

		reply := ""
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// логируем от кого какое сообщение пришло
		//	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		userNick := update.Message.From.UserName
		userStatus := update.Message.Text
		//userState := userNick + " - " + userStatus // для отладки

		//sendUserState := tgbotapi.NewMessage(update.Message.Chat.ID, userState)

		if strings.Contains(userStatus, "/") {
			switch update.Message.Command() {
			case "status_sect": // пришлет статус всех в меню бота
				reply = fmt.Sprintln(listUsers)
				replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
				bot.Send(replyMsg)
			case "send": // пошлет в чат chatID статус
				go sendNotifications(bot)
			case "read":
				listUsers = readFromFile(botconfd.FileJson)
			case "write":
				writeToFile(listUsers)
				fmt.Println("write Ok", listUsers)
			}
		} else {
			if _, ok := users[userNick]; ok {
				users[userNick] = userStatus
				for key1, val1 := range users {
					for key2, val2 := range names {
						if key1 == val2 {
							listUsers[key2] = val1
						}
					}
				}
				writeToFile(listUsers)
			}
		}
	}
}
func sendNotifications(bot *tgbotapi.BotAPI) {
	var stringMap string
	stringMap = mapToList(listUsers)
	bot.Send(tgbotapi.NewMessage(botconfd.ChatID, stringMap))
}

func sendReport(bot *tgbotapi.BotAPI) {
	for {
		t := time.Now()
		ts := t.Format(time.UnixDate)
		tsArr := strings.Split(ts, " ")
		tts := tsArr[3]
		ttsArr := strings.Split(tts, ":")
		hour := ttsArr[0]
		minute := ttsArr[1]
		if hour == botconfd.SendHour1 && minute == botconfd.SendMinute1 {
			sendNotifications(bot)
			time.Sleep(time.Minute * 1)
		}
		if hour == botconfd.SendHour2 && minute == botconfd.SendMinute2 {
			sendNotifications(bot)
			time.Sleep(time.Minute * 1)
		}
	}
}

func mapToList(m map[string]interface{}) string {
	var str string
	var strAll string
	for key, val := range m {
		str = fmt.Sprintf("%s - %s", key, val)
		strAll = strAll + "\n" + str
	}
	return strAll
}

func readFromFile(s string) map[string]interface{} { //принимает имя файла, возвращает map
	file, err1 := os.Open(s)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer file.Close()
	jsonString, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		log.Fatal(err2)
	}
	var str nameFromJson
	err3 := json.Unmarshal(jsonString, &str)
	if err3 != nil {
		log.Fatal(err3)
	}
	m := structs.Map(str)
	return m
}

func writeToFile(m map[string]interface{}) {
	var str nameFromJson
	err1 := mapstructure.Decode(m, &str)
	if err1 != nil {
		panic(err1)
	}
	jsn, err2 := json.Marshal(str)
	if err2 != nil {
		panic(err2)
	}
	err3 := ioutil.WriteFile(botconfd.FileJson, jsn, 0644)
	if err3 != nil {
		panic(err3)
	}
}
