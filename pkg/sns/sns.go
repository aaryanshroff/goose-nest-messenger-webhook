// Package sns provides a simple wrapper around the AWS SNS client
package sns

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSMessage struct {
	Body        string `json:"Body"`
	RecipientId string `json:"RecipientId"`
}

func PublishToSNS(topicArn string, message string) (*sns.PublishOutput, error) {
	// Create the SNS client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sns.New(sess)

	// Publish the message
	result, err := svc.Publish(&sns.PublishInput{
		Message:  &message,
		TopicArn: &topicArn,
	})

	return result, err
}

// GetMessaageFromSNSEvent returns the first message from an SNS event
func GetMessageFromSNSEvent(event events.SNSEvent) (SNSMessage, error) {
	message := SNSMessage{}

	messageJSONString := event.Records[0].SNS.Message
	err := json.Unmarshal([]byte(messageJSONString), &message)
	if err != nil {
		errWithCtx := fmt.Errorf("unmarshalling SNS message: %w", err)
		return message, errWithCtx
	}

	return message, nil
}
