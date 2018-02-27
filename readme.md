# Objective:
Create a service that helps to be on top of the car buying market.
Service scans the pages and informs user as soon as new car appears.
As well as stores data for further analysis.

Interaction happens through Telegram application.
User can interact by registering and sending commands.

***

### Commands supported:
- /register <password>
- /queries - display saved serch queries
- /add <link> - adds new query
- /delete <link id>
- /check - for new ads
- /enable - enable auto updates
- /disable
- /list-all-commands
			
### Basic flow:
1. User sends check command
2. Telegram-bot Lambda tries to download user data from Dynamo DB.
3. If relevant data is present. With each saved query invokes scrapper Lambdas.
4. Scrapper - scrapes all results and sends back.
5. Telegram-bot then replies back with new cars to user and stores them in DB for later comparison.

#### Services used:
- AWS Lambda
- AWS CloudWatch Events
- AWS Dynamo DB

#### Also:
- ngrok
- postman
