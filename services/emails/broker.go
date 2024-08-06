package emails

import (
	"net/smtp"

	"context"
	"errors"
	"os"
	"strconv"

	"github.com/flaviofrancisco/vagasprajr-api-v2/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailSettings struct {
	Host string    `json:"host"`
	Port int       `json:"port"`
	Auth EmailAuth `json:"auth"`
	From string    `json:"from"`
	To   string    `json:"to"`
}

type EmailAuth struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

func GetEmailSettings() (EmailSettings, error) {

	mongodb_database := os.Getenv("MONGODB_DATABASE")
	client, err := models.Connect()

	// Ensure the client connection is closed once the function completes
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return EmailSettings{}, err
	}

	db := client.Database(mongodb_database)

	// Get the email settings where _id = 'mail'
	filter := bson.D{{"_id", "mail"}}
	var result EmailSettings
	err = db.Collection("settings").FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return EmailSettings{}, errors.New("Configurações de email não encontradas")
		}
	}

	return result, nil
}

func SendEmail(from string, to []string, subject string, body string) error {

	// Get email settings
	settings, err := GetEmailSettings()

	if err != nil {
		return err
	}

	if from == "" {
		from = settings.From
	}

	if len(to) == 0 {
		to = append(to, settings.To)
	}

	smtpHost := settings.Host
	smtpPort := settings.Port
	auth := smtp.PlainAuth("", settings.Auth.User, settings.Auth.Pass, smtpHost)

	// Prepare email body
	msg := "From: " + from + "\n" +
		"To: " + to[0] + "\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"" + "\n\n" +
		body

	err = smtp.SendMail(smtpHost+":"+strconv.Itoa(smtpPort), auth, from, to, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
