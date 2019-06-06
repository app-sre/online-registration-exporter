package onlinereg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	APIUrl   string
	APIUser  string
	APIToken string
}

type onlineRegClient struct {
	config Config
	client *http.Client
}

func NewClient(c Config) *onlineRegClient {
	orc := &onlineRegClient{
		config: c,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
	return orc
}

type Capacity struct {
	Plan struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		Type              string `json:"type"`
		DisplayName       string `json:"display_name"`
		IsHidden          bool   `json:"is_hidden"`
		SubscriberLimit   int    `json:"subscriber_limit"`
		CapacityConsumed  int    `json:"capacity_consumed"`
		CapacityRemaining int    `json:"capacity_remaining"`
	} `json:"plan"`
}

func (orc onlineRegClient) GetPlanCapacity(plan string) (Capacity, error) {
	c := Capacity{}

	url, err := url.Parse(fmt.Sprintf("%s/plans/%s/capacity", orc.config.APIUrl, plan))
	if err != nil {
		return c, err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return c, err
	}

	req.Header.Set("Authorization", "Bearer "+orc.config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("authorization_username", orc.config.APIUser)
	req.URL.RawQuery = q.Encode()

	res, err := orc.client.Do(req)
	if err != nil {
		return c, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return c, err
	}

	if err := json.Unmarshal(body, &c); err != nil {
		return c, err
	}

	return c, nil
}
