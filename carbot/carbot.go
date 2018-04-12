package carbot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Carbot accepts messages and sends back response to the client.
// Needed:
// -	Bot token
// -	Dynamo DB settings (region, table name, primary key)
type Carbot struct {
	telegram *tgbotapi.BotAPI
	database *DynamoAPI
	scrapper *ScrapperAPI
	regPass  string
	message  string
	userID   int64
}

// NewCarbot is main object of this car-ad-helper to work with.
func NewCarbot(botToken, dbRegion, dbTableName, dbPrimaryKey, regPass, scrapperURL string) (*Carbot, error) {
	cb := &Carbot{}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	cb.telegram = bot

	db := NewDynamoAPI(dbRegion, dbTableName, dbPrimaryKey)
	cb.database = db

	scr := NewScrapperAPI(scrapperURL)
	cb.scrapper = scr
	cb.regPass = regPass

	return cb, nil
}

func (c *Carbot) UnmarshalData(data string) {
	var update tgbotapi.Update
	err := json.Unmarshal([]byte(data), &update)
	if err != nil {
		log.Printf("Error: %v \nReceived Data: %s", err, data)
		// return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 400}, nil
	}
	//log.Printf("Message from ID [%d] received: %s", update.Message.Chat.ID, update.Message.Text)

	c.userID = update.Message.Chat.ID
	c.message = update.Message.Text
}

// Respond selects correct response and sends it back.
// Accepts Json received through endpoint
func (c *Carbot) Respond(data string) error {
	c.UnmarshalData(data) //unmarshal json

	dbuser, err := c.database.Retrieve(c.userID) //Get User data from DynamoDB
	if err != nil {
		log.Println(err)
		return err
	}

	cmd := strings.Split(c.message, " ")
	if dbuser == nil { //User not registered
		if cmd[0] != "/register" {
			c.respondText("Please register with service. Send message \"/register <password>\"")
		} else {
			if cmd[1] != c.regPass {
				c.respondText("Password provided is wrong. Please contact service administrator.")
			} else {
				c.respondUserRegister()
			}
		}
	} else {
		switch cmd[0] { //User registered. All Good.
		case "/register":
			c.respondKeyboard("User already registered.")
		case "/queries":
			c.respondQueryShowSaved(dbuser)
		case "/add":
			c.respondQueryAdd(cmd[1], dbuser)
		case "/delete":
			c.respondQueryDelete(cmd[1], dbuser)
		case "/check":
			c.respondUserCheck("Check finished. No new cars.", dbuser)
		case "/enable":
			c.respondAutoUpdateEnable(dbuser)
		case "/disable":
			c.respondAutoUpdateDisable(dbuser)
		case "/list-all-commands":
			c.respondText("Available commands:\n/add <link> - add new search\n/queries - see saved search queries.\n/delete <No.> - delete search\n/check - scan for new ads\n/enable - to turn on auto search\n/disable - to turn off auto search")
		default:
			c.respondKeyboard("Unrecognized command.")
		}
	}
	return nil
}

func (c *Carbot) respondUserRegister() {
	user := NewDBUser(c.userID)
	err := c.database.Store(user)
	if err != nil {
		c.respondKeyboard("Registration failed.")
	} else {
		c.respondKeyboard("Registration successful.\n To add new search. Please send us link used for search.\n Currently supported: Autoplius.lt")
	}
}

// CheckAll checks and informs every user from database if new cars were added.
// Invoked by /check from telegram and scheduled event on AWS cloudwatch
func (c *Carbot) CheckAll( /*input string*/ ) error {
	//cmd := strings.Split(input, " ")
	//if cmd[0] == "/check_all" && cmd[1] == checkAllPassword {
	// if input == os.Getenv("ALL_PASS") {
	users, err := c.database.RetrieveAll()
	if err != nil {
		// Check logs for more data with this id...
		return err
	}

	for _, user := range users {
		if user.AutoUpdate && len(user.Queries) != 0 {
			c.respondUserCheck("", &user)
		}
	}
	log.Println("All users checked")
	// }
	return nil
}

//================================================
//					Queries
//================================================

// Prints out saved queries to the client.
// Example:
// 1) https://www.autoplius.com/...
// 2) https://www.autoplius.com/...
func (c *Carbot) respondQueryShowSaved(user *DBUser) {
	msgText := ""
	if len(user.Queries) < 1 {
		msgText = "You have no queries saved. Please add query using /add <link>"
	} else {
		for i, v := range user.Queries {
			msgText += fmt.Sprintf("%d) %s\n", i+1, v)
		}
	}
	c.respondText(msgText)
}

// Add Search link to saved Queries. Thse are used as sources to scrape data from.
func (c *Carbot) respondQueryAdd(link string, user *DBUser) {
	msgText := ""
	if VerifySearchLink(link) == false {
		msgText = "Link could not be verified. Please inspect the link and try again."
	} else {
		if user.QueryExist(link) { //If already exist
			msgText = "Search already saved."
		} else { // Otherwise store all cars so we can check later for newly appeared ones.
			resp := c.scrapper.Invoke(link)
			user.CarsAdd(resp.Results)
			user.QueryAdd(link)
			err := c.database.Store(user)

			if err != nil {
				msgText = "Failed saving query."
			} else {
				msgText = "Search saved."
			}
		}
	}
	c.respondText(msgText)
}

// Delete Saved Query by ID.
func (c *Carbot) respondQueryDelete(command string, user *DBUser) {
	msgText := ""
	id, err := strconv.Atoi(command)
	id--

	if err != nil {
		msgText = "Invalid number. Try integers. For example: 0"
	} else if id >= 0 && id <= len(user.Queries) { //If already exist
		user.QueryDeleteID(id)
		c.database.Store(user)
		msgText = "Search deleted successfully."
	} else {
		msgText = "Search could not be found. Please type /queries to see all saved searches"
	}
	c.respondText(msgText)
}

// Checks for udates per specific user.
// defaultMsg is sent if no new cars were found. If left empty, no message will be sent.
func (c *Carbot) respondUserCheck(defaultMsg string, user *DBUser) {
	c.userID = user.ID

	if len(user.Queries) == 0 {
		c.respondText("No searches saved. Please add one using /add <link>")
	}

	// Collect results from all requests.
	allCars := map[string]bool{}
	for _, v := range user.Queries {
		resp := c.scrapper.Invoke(v)
		for kk := range resp.Results {
			allCars[kk] = true
		}
	}
	log.Printf("User %d has %d cars scrapped", c.userID, len(allCars))

	// Find and send new cars.
	newCount := 0
	for newCar := range allCars {
		if user.Cars[newCar] == false && user.NotInterested[newCar] == false {
			newCount++
			c.respondText(newCar)
			defaultMsg = ""
			log.Println("Message sent to user")
		}
	}

	// Store all
	user.CarsAdd(allCars)
	c.database.Store(user)

	if defaultMsg != "" {
		c.respondText(defaultMsg)
	}
}

//================================================
//					Auto Updates
//================================================

func (c *Carbot) respondAutoUpdateEnable(user *DBUser) {
	//Dont allow if user has no saved searches.
	if len(user.Queries) < 1 {
		c.respondText("Please add query using /add <link> before enabling auto updates.")
	} else {
		user.AutoUpdate = true
		c.database.Store(user)
		c.respondText("Auto updates enabled")
	}
}

func (c *Carbot) respondAutoUpdateDisable(user *DBUser) {
	user.AutoUpdate = false
	c.database.Store(user)
	c.respondText("Auto updates disabled")
}

func (c *Carbot) respondText(msgText string) {
	msg := tgbotapi.NewMessage(c.userID, msgText)
	c.telegram.Send(msg)
}

func (c *Carbot) respondKeyboard(msgText string) {
	msg := tgbotapi.NewMessage(c.userID, msgText)
	msg.ReplyMarkup = &tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.KeyboardButton{Text: "/list-all-commands"}},
			{tgbotapi.KeyboardButton{Text: "/queries"}, tgbotapi.KeyboardButton{Text: "/check"}},
			{tgbotapi.KeyboardButton{Text: "/enable"}, tgbotapi.KeyboardButton{Text: "/disable"}},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	c.telegram.Send(msg)
}
