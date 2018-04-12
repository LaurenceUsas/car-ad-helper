package carbot_test

import (
	"reflect"
	"testing"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func TestNewCarbot(t *testing.T) {
	type args struct {
		botToken     string
		dbRegion     string
		dbTableName  string
		dbPrimaryKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *Carbot
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCarbot(tt.args.botToken, tt.args.dbRegion, tt.args.dbTableName, tt.args.dbPrimaryKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCarbot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCarbot() = %v, want %v", got, tt.want)
			}
		})
	}
}

//Okay Json
//Bad json
//---
//No userID
//No message
func TestCarbot_parseIncoming(t *testing.T) {
	type fields struct {
		telegram *tgbotapi.BotAPI
		database *DynamoAPI
		message  string
		userID   int64
	}
	type args struct {
		data string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Carbot{
				telegram: tt.fields.telegram,
				database: tt.fields.database,
				message:  tt.fields.message,
				userID:   tt.fields.userID,
			}
			c.parseIncoming(tt.args.data)
		})
	}
}
