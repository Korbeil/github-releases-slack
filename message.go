package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"text/template"
	"regexp"
	"strings"
)

type MessageTemplateData struct {
	URL       string
	FullName  string
	Version   string
	Changelog string
	Channel   string
}

type MessageResponse struct {
	OK bool `json:"ok"`
}

func messageFromRequest(request Request) ([]byte, error) {
	// get CHANGELOG contents
	contents, err := getChangelogContents(request.Payload)
	if err != nil {
		return nil, err
	}
	contents = getStringInBetween(contents, request.Payload)
	contents = strings.Replace(contents, `\`, `|`, -1)

	template := messageTemplateFromPayloadForChannel(request.Payload, request.Channel, contents)

	if request.Payload.RefType != "tag" {
		var err error
		return nil, err
	}

	return messageFromTemplate(template)
}

func getChangelogContents(payload Payload) (string, error) {
	var url = "https://raw.githubusercontent.com/" + payload.Repository.FullName + "/" + payload.Ref + "/CHANGELOG.md"
	var resp, err = http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(contents), nil
	}
	return "", err
}

func getStringInBetween(str string, payload Payload) (result string) {
	r := regexp.MustCompile(`(?s)(##\s\[` + payload.Ref + `\].*?)(##\s\[)`)
	res := r.FindStringSubmatch(str)
	return strings.TrimSpace(res[1])
}

func messageTemplateFromPayloadForChannel(payload Payload, channel string, changelog string) MessageTemplateData {
	return MessageTemplateData{
		payload.Repository.URL,
		payload.Repository.FullName,
		payload.Ref,
		changelog,
		channel,
	}
}

func messageFromTemplate(data MessageTemplateData) ([]byte, error) {
	t, _ := template.ParseFiles("message.json")

	var message bytes.Buffer
	if err := t.Execute(&message, data); err != nil {
		return nil, err
	}

	return message.Bytes(), nil
}

func postMessageToSlack(message []byte, token string) (*MessageResponse, error) {
	url := "https://slack.com/api/chat.postMessage"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var response MessageResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}
