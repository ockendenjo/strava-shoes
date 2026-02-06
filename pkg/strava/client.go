package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func NewClient(ssmClient *ssm.Client, httpClient *http.Client) *Client {
	psClient := newParamsClient(ssmClient)
	return &Client{psClient: psClient, httpClient: httpClient}
}

type Client struct {
	psClient   *paramsClient
	httpClient *http.Client
	params     *StravaParams
}

func (c *Client) Authorize(ctx context.Context, code string) error {
	gotParams, err := c.psClient.GetParams(ctx)
	if err != nil {
		return err
	}
	c.params = &gotParams

	u, err := url.Parse("https://www.strava.com/oauth/token")
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	q.Set("client_id", c.params.ClientId)
	q.Set("client_secret", c.params.ClientSecret)
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), nil)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return HttpStatusError{StatusCode: res.StatusCode, Body: string(bytes)}
	}

	var refreshRes refreshResponse
	err = json.Unmarshal(bytes, &refreshRes)
	if err != nil {
		return err
	}

	c.params.AccessToken = refreshRes.AccessToken
	c.params.RefreshToken = refreshRes.RefreshToken
	c.params.ExpiryTime = refreshRes.ExpiryTime

	//Set SSM params
	err = c.psClient.SetRefreshedParams(ctx, refreshRes.RefreshToken, refreshRes.AccessToken, refreshRes.ExpiryTime)
	return err
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	if c.params == nil {
		gotParams, err := c.psClient.GetParams(ctx)
		if err != nil {
			return "", err
		}
		c.params = &gotParams
	}

	if c.params.ExpiryTime > time.Now().Unix() {
		return c.params.AccessToken, nil
	}

	//Need to refresh
	u, err := url.Parse("https://www.strava.com/oauth/token")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client_id", c.params.ClientId)
	q.Set("client_secret", c.params.ClientSecret)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", c.params.RefreshToken)
	u.RawQuery = q.Encode()
	fmt.Println(u)

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), nil)
	if err != nil {
		return "", err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", HttpStatusError{StatusCode: res.StatusCode, Body: string(bytes)}
	}

	var refreshRes refreshResponse
	err = json.Unmarshal(bytes, &refreshRes)
	if err != nil {
		return "", err
	}

	c.params.AccessToken = refreshRes.AccessToken
	c.params.RefreshToken = refreshRes.RefreshToken
	c.params.ExpiryTime = refreshRes.ExpiryTime

	//Set SSM params
	err = c.psClient.SetRefreshedParams(ctx, refreshRes.RefreshToken, refreshRes.AccessToken, refreshRes.ExpiryTime)
	if err != nil {
		return "", err
	}
	return refreshRes.AccessToken, nil
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiryTime   int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) GetActivities(ctx context.Context, page int) ([]Activity, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	stravaUrl := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?page=%d", page)
	req, err := http.NewRequestWithContext(ctx, "GET", stravaUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, HttpStatusError{StatusCode: res.StatusCode, Body: string(bytes)}
	}

	var activities []Activity
	err = json.Unmarshal(bytes, &activities)
	return activities, err
}

type Activity struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SportType string `json:"sport_type"`
	GearID    string `json:"gear_id"`
}

type HttpStatusError struct {
	StatusCode int
	Body       string
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("HTTP request returned status code %d: %s", e.StatusCode, e.Body)
}
