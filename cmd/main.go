package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aaryanshroff/rentals-bot-messenger-webhook/pkg/messenger"
	"github.com/aaryanshroff/rentals-bot-messenger-webhook/pkg/sns"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event json.RawMessage) (events.LambdaFunctionURLResponse, error) {
	var request events.LambdaFunctionURLRequest
	var snsEvent events.SNSEvent

	// SNS event
	if err := json.Unmarshal(event, &snsEvent); err == nil && len(snsEvent.Records) > 0 {
		return handleSNSEvent(ctx, snsEvent)
	}

	// Messenger event
	if err := json.Unmarshal(event, &request); err == nil && request.RequestContext.HTTP.Method != "" {
		switch request.RequestContext.HTTP.Method {
		case "GET":
			return handleGet(ctx, request)
		case "POST":
			return handlePost(ctx, request)
		default:
			return events.LambdaFunctionURLResponse{StatusCode: 405}, nil
		}
	}

	return events.LambdaFunctionURLResponse{StatusCode: 400}, nil
}

func handleSNSEvent(ctx context.Context, event events.SNSEvent) (events.LambdaFunctionURLResponse, error) {
	log.Printf("SNS event: %v", event)

	message, err := sns.GetMessageFromSNSEvent(event)
	if err != nil {
		log.Printf("Error getting SNS message: %v", err)
		return events.LambdaFunctionURLResponse{StatusCode: 400}, err
	}

	log.Printf("SNS message: %+v", message)

	err = messenger.SendMessage(message.RecipientId, message.Body)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return events.LambdaFunctionURLResponse{StatusCode: 400}, err
	}

	return events.LambdaFunctionURLResponse{StatusCode: 200}, nil

}

// Messenger verification request
// https://developers.facebook.com/docs/messenger-platform/webhooks#verification-requests
func handleGet(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	mode := request.QueryStringParameters["hub.mode"]
	token := request.QueryStringParameters["hub.verify_token"]
	challenge := request.QueryStringParameters["hub.challenge"]

	verify_token := os.Getenv("VERIFY_TOKEN")

	if token != verify_token {
		return events.LambdaFunctionURLResponse{StatusCode: 403}, nil
	}

	if mode == "subscribe" {
		log.Printf("Webhook verified")
		return events.LambdaFunctionURLResponse{StatusCode: 200, Body: challenge}, nil
	}

	return events.LambdaFunctionURLResponse{StatusCode: 400}, nil
}

// Messenger event notification
// https://developers.facebook.com/docs/messenger-platform/webhooks#event-notifications
func handlePost(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {

	event := messenger.Event{}

	body := request.Body
	err := json.Unmarshal([]byte(body), &event)
	if err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return events.LambdaFunctionURLResponse{StatusCode: 400}, err
	}

	switch event.Object {
	case "page":
		for _, entry := range event.Entry {
			for _, messagingEvent := range entry.Messaging {
				log.Printf("Message received: %v", messagingEvent.Message.Text)
				topic := os.Getenv("SNS_TOPIC")
				snsMessage := sns.SNSMessage{
					RecipientId: messagingEvent.Sender.ID,
					Body:        messagingEvent.Message.Text,
				}
				snsMessageBytes, err := json.Marshal(snsMessage)
				if err != nil {
					log.Printf("Error marshalling SNS message: %v", err)
					return events.LambdaFunctionURLResponse{StatusCode: 500}, err
				}
				_, err = sns.PublishToSNS(topic, string(snsMessageBytes))
				if err != nil {
					log.Printf("Error publishing to SNS: %v", err)
					return events.LambdaFunctionURLResponse{StatusCode: 500}, err
				}
			}
		}
	default: // Unsupported event type
		return events.LambdaFunctionURLResponse{StatusCode: 400}, nil
	}

	return events.LambdaFunctionURLResponse{StatusCode: 200}, nil
}
