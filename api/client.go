package api

import (
	"bytes"
	"dnscheck/dbs"
	"encoding/json"
	"fmt"
	"net/http"
)

type ApiV1Client struct {
	serverUrl string
	secret    string
	clientId  string
}

func NewApiV1Client(url, secret, id string) *ApiV1Client {

	clientId := id
	if clientId == "" {
		clientId = "new"
	}
	s := &ApiV1Client{
		serverUrl: url,
		secret:    secret,
		clientId:  clientId,
	}

	return s
}

func (s *ApiV1Client) GetConfig() (*dbs.ClientConfig, error) {

	// Post request to server for blind signing the message
	request, err := http.NewRequest("GET", fmt.Sprintf("%s%s%s", s.serverUrl, withPrefix("/config/"), s.clientId), nil)
	if err != nil {
		return nil, err
	}
	token, err := issueJwt(s.clientId, []byte(s.secret))
	if err != nil {
		return nil, fmt.Errorf("Failed to issue jwt: %s", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Check HTTP status
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server returned not successful response %s", response.Status)
	}

	var conf dbs.ClientConfig
	if err := json.NewDecoder(response.Body).Decode(&conf); err != nil {
		return nil, fmt.Errorf("Failed to read response body: %s", err)
	}

	return &conf, nil
}

func (s *ApiV1Client) SetResults(info dbs.ResultInfo) error {

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", s.serverUrl, withPrefix("/results")), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	token, err := issueJwt(s.clientId, []byte(s.secret))
	if err != nil {
		return fmt.Errorf("Failed to issue jwt: %s", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check HTTP status
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned not successful response %s", response.Status)
	}

	return nil
}
