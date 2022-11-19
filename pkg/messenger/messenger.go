// Pacakage messenger provides a simple interface for sending messages to Facebook Messenger
package messenger

import (
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"os"
)

type Event struct {
	Object string `json:"object"`
	Entry  []struct {
		ID        string `json:"id"`
		Time      int64  `json:"time"`
		Messaging []struct {
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Recipient struct {
				ID string `json:"id"`
			} `json:"recipient"`
			Timestamp int64 `json:"timestamp"`
			Message   struct {
				Mid  string `json:"mid"`
				Text string `json:"text"`
			}
		} `json:"messaging"`
	} `json:"entry"`
}

type Message struct {
	Text string `json:"text"`
}

type Recipient struct {
	ID string `json:"id"`
}

func SendMessage(recipient_id string, message string) error {
	// Create the message
	message_struct := Message{
		Text: message,
	}
	message_json, err := json.Marshal(message_struct)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	message_json_string := string(message_json)

	// Create the recipient
	recipient_struct := Recipient{
		ID: recipient_id,
	}
	recipient_json, err := json.Marshal(recipient_struct)
	if err != nil {
		return fmt.Errorf("marshal recipient: %w", err)
	}
	recipient_json_string := string(recipient_json)

	// Create the request
	page_id := os.Getenv("PAGE_ID")
	if page_id == "" {
		return fmt.Errorf("PAGE_ID not set")
	}

	page_access_token := os.Getenv("PAGE_ACCESS_TOKEN")
	if page_access_token == "" {
		return fmt.Errorf("PAGE_ACCESS_TOKEN not set")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v15.0/%s/messages", page_id)

	messaging_type := "RESPONSE"

	params := urlpkg.Values{
		"access_token":   {page_access_token},
		"message":        {message_json_string},
		"messaging_type": {messaging_type},
		"recipient":      {recipient_json_string},
	}

	// Send the request
	resp, err := http.PostForm(url, params)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("sending message: %s", resp.Body)
	}

	return nil
}
