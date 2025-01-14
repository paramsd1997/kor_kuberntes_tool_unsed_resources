package kor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type SendMessageToSlack interface {
	SendToSlack(slackOpts SlackOpts, outputBuffer string) error
}

type SlackOpts struct {
	WebhookURL string
	Channel    string
	Token      string
}

type SlackMessage struct {
}

func SendToSlack(sm SendMessageToSlack, slackOpts SlackOpts, outputBuffer string) error {
	return sm.SendToSlack(slackOpts, outputBuffer)
}

func (sm SlackMessage) SendToSlack(slackOpts SlackOpts, outputBuffer string) error {
	if slackOpts.WebhookURL != "" {
		payload := []byte(`{"text": "` + outputBuffer + `"}`)
		_, err := http.Post(slackOpts.WebhookURL, "application/json", bytes.NewBuffer(payload))

		if err != nil {
			return err
		}
		return nil
	} else if slackOpts.Channel != "" && slackOpts.Token != "" {
		fmt.Printf("Sending message to Slack channel %s...", slackOpts.Channel)
		outputFilePath, _ := writeOutputToFile(outputBuffer)

		var formData bytes.Buffer
		writer := multipart.NewWriter(&formData)

		fileWriter, err := writer.CreateFormFile("file", outputFilePath)
		if err != nil {
			return err
		}
		file, err := os.Open(outputFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(fileWriter, file)
		if err != nil {
			return err
		}

		if err := writer.WriteField("channels", slackOpts.Channel); err != nil {
			return err
		}

		writer.Close()

		req, err := http.NewRequest("POST", "https://slack.com/api/files.upload", &formData)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+slackOpts.Token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("slack API returned non-OK status code: %d", resp.StatusCode)
		}

		return nil
	} else {
		return errors.New("SlackOpts must contain either WebhookURL or Channel and Token")
	}
}

func writeOutputToFile(outputBuffer string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %v", err)
	}

	outputFileName := "kor-scan-results.txt"
	outputFilePath := filepath.Join(homeDir, outputFileName)

	file, err := os.Create(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(outputBuffer)
	if err != nil {
		return "", fmt.Errorf("failed to write output to file: %v", err)
	}

	return outputFilePath, nil
}
