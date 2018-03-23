package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/LaurenceUsas/car-ad-helper/api-dynamo"
	"github.com/LaurenceUsas/car-ad-helper/api-scrapper"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	// Dynamo DB
	dbRegion     = "eu-west-1"
	dbTableName  = "car-ad-helper"
	dbPrimaryKey = "userID"
)

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.Body == "" {
		log.Print("Checking All")
		checkAll( /*os.Getenv("ALL_PASS")*/ )
	} else {
		log.Print("Processing Command")
		var update tgbotapi.Update
		err := json.Unmarshal([]byte(request.Body), &update)
		if err != nil { //TODO Validate JSON? Like checking if all data that will be used later exists.
			log.Printf("Error: %v \nMessage: %s", err, request.Body)
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
		}
		log.Printf("Message from ID [%d] received: %s", update.Message.Chat.ID, update.Message.Text)

		sendResponse(update.Message)
	}
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func checkAll( /*input string*/ ) {
	//cmd := strings.Split(input, " ")
	//if cmd[0] == "/check_all" && cmd[1] == checkAllPassword {
	// if input == os.Getenv("ALL_PASS") {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		return
	}

	db := cahdynamo.NewDynamoAPI(dbRegion, dbTableName, dbPrimaryKey)
	users, err := db.RetrieveAll()
	if err != nil {
		return
	}

	for _, user := range users {
		if user.AutoUpdate && len(user.Queries) != 0 {
			respondCheckUser("", &user, bot, db)
		}
	}
	log.Println("All users checked")
	// }
}

func sendResponse(message *tgbotapi.Message) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	db := cahdynamo.NewDynamoAPI(dbRegion, dbTableName, dbPrimaryKey)
	user, err := db.Retrieve(message.Chat.ID) //Get User data from DynamoDB
	if err != nil {
		log.Println(err)
	}

	msgText := ""
	addKeyboard := false
	cmd := strings.Split(message.Text, " ")
	if cmd[0] == "/register" {
		if cmd[1] == os.Getenv("REG_PASS") {
			if user != nil {
				msgText = "User already registered."
			} else {
				// Add to DB
				user := cahdynamo.NewDBUser(message.Chat.ID)
				err := db.Store(user)

				if err != nil {
					msgText = "Registration failed."
				} else {
					msgText = "Registration successful.\n To add new search. Please send us link used for search.\n Currently supported: Autoplius.lt" // TODO Explain how to use.
					addKeyboard = true
				}
			}

		}
	} else if user == nil { // If sending anything but not registred. Tell to register
		msgText = "Please register with service. Send message \"/register <password>\""
	} else { // Otherwise look for command.
		switch cmd[0] {
		case "/queries":
			respondList(user, bot, db)
		case "/add":
			respondAddLink(cmd[1], user, bot, db)
		case "/delete":
			respondDeleteLink(cmd[1], user, bot, db)
		case "/check":
			respondCheckUser("Check finished. No new cars.", user, bot, db)
		case "/enable":
			//Dont allow if user has no saved searches.
			if len(user.Queries) < 1 {
				msgText = "Please add query using /add <link> before enabling auto updates."
			} else {
				user.AutoUpdate = true
				db.Store(user)
				msgText = "Auto updates enabled"
			}
		case "/disable":
			user.AutoUpdate = false
			db.Store(user)
			msgText = "Auto updates disabled"
		case "/list-all-commands":
			msgText = "Available commands:\n/add <link> - add new search\n/queries - see saved search queries.\n/delete <No.> - delete search\n/check - scan for new ads\n/enable - to turn on auto search\n/disable - to turn off auto search"
		default:
			msgText = "Unrecognized command."
			addKeyboard = true
		}
	}

	if msgText != "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		if addKeyboard {
			msg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
				Keyboard: [][]tgbotapi.KeyboardButton{
					{tgbotapi.KeyboardButton{Text: "/list-all-commands"}},
					{tgbotapi.KeyboardButton{Text: "/queries"}, tgbotapi.KeyboardButton{Text: "/check"}},
					{tgbotapi.KeyboardButton{Text: "/enable"}, tgbotapi.KeyboardButton{Text: "/disable"}},
				},
				ResizeKeyboard:  true,
				OneTimeKeyboard: true,
			}
		}
		bot.Send(msg)
	}
}

func respondList(user *cahdynamo.DBUser, bot *tgbotapi.BotAPI, db *cahdynamo.DynamoAPI) {
	msgText := ""
	if len(user.Queries) < 1 {
		msgText = "You have no queries saved. Please add query using /add <link>"
	} else {
		for i, v := range user.Queries {
			msgText += fmt.Sprintf("%d) %s\n", i+1, v)
		}
	}
	msg := tgbotapi.NewMessage(user.ID, msgText)
	bot.Send(msg)
}

func respondAddLink(link string, user *cahdynamo.DBUser, bot *tgbotapi.BotAPI, db *cahdynamo.DynamoAPI) {
	msgText := ""
	if verifySearchLink(link) == false {
		msgText = "Link could not be verified. Please inspect the link and try again."
	} else {
		if user.QueryExist(link) { //If already exist
			msgText = "Search already saved."
		} else {
			// Store all cars so we can check later for newly appeared ones.
			// Store Query
			scraperURL := os.Getenv("URL_SCRAPER")
			scrapper := scrapper.NewScrapperAPI(scraperURL)
			resp := scrapper.Invoke(link)

			user.CarsAdd(resp.Results)
			user.QueryAdd(link)
			err := db.Store(user)

			if err != nil {
				msgText = "Failed saving query."
			} else {
				msgText = "Search saved."
			}
		}
	}
	msg := tgbotapi.NewMessage(user.ID, msgText)
	bot.Send(msg)
}

func respondDeleteLink(command string, user *cahdynamo.DBUser, bot *tgbotapi.BotAPI, db *cahdynamo.DynamoAPI) {
	msgText := ""
	id, err := strconv.Atoi(command)
	id--

	if err != nil {
		msgText = "Invalid number. Try integers. For example: 0"
	} else if id >= 0 && id <= len(user.Queries) { //If already exist
		user.QueryDeleteID(id)
		db.Store(user)
		msgText = "Search deleted successfully."
	} else {
		msgText = "Search could not be found. Please type /queries to see all saved searches"
	}
	msg := tgbotapi.NewMessage(user.ID, msgText)
	bot.Send(msg)
}

func respondCheckUser(defaultMsg string, user *cahdynamo.DBUser, bot *tgbotapi.BotAPI, db *cahdynamo.DynamoAPI) {
	if len(user.Queries) == 0 {
		msg := tgbotapi.NewMessage(user.ID, "No searches saved. Please add one using /add <link>")
		bot.Send(msg)
	}

	allCars := map[string]bool{}
	// Collect results from all requests.
	scraperURL := os.Getenv("URL_SCRAPER")
	scrapper := scrapper.NewScrapperAPI(scraperURL)

	for _, v := range user.Queries {
		resp := scrapper.Invoke(v)

		for kk := range resp.Results {
			allCars[kk] = true
		}
	}
	log.Printf("User %d has %d cars scrapped", user.ID, len(allCars))

	// Find and send new cars.
	newcars := 0
	for k := range allCars {
		if user.Cars[k] == false && user.NotInterested[k] == false {
			newcars++
			msg := tgbotapi.NewMessage(user.ID, k)
			bot.Send(msg)
			defaultMsg = ""
			log.Println("Message sent to user")
		}
	}

	// Store all
	user.CarsAdd(allCars)
	db.Store(user)

	if defaultMsg != "" {
		msg := tgbotapi.NewMessage(user.ID, defaultMsg)
		bot.Send(msg)
	}
}

func verifySearchLink(url string) bool {
	if !strings.Contains(url, "https://autoplius.lt/") {
		return false
	}
	resp, _ := http.Get(url)
	if resp.StatusCode != 200 {
		return false
	}
	log.Println("Link verified!")
	return true
}
