package strava

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/jsii-runtime-go"
	"strconv"
	"strings"
)

const (
	prefix          = "/strava/"
	keyAccessToken  = "accessToken"
	keyRefreshToken = "refreshToken"
	keyExpiryTime   = "expiryTime"
)

func newParamsClient(client *ssm.Client) *paramsClient {
	return &paramsClient{ssmClient: client}
}

type paramsClient struct {
	ssmClient *ssm.Client
}

func (c *paramsClient) GetParams(ctx context.Context) (StravaParams, error) {

	res, err := c.ssmClient.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:           jsii.String(prefix),
		Recursive:      jsii.Bool(true),
		WithDecryption: jsii.Bool(true),
	})
	if err != nil {
		return StravaParams{}, err
	}

	sp := StravaParams{}

	for _, parameter := range res.Parameters {
		key := strings.Replace(*parameter.Name, prefix, "", -1)
		switch key {
		case "clientId":
			sp.ClientId = *parameter.Value
		case "clientSecret":
			sp.ClientSecret = *parameter.Value
		case keyAccessToken:
			sp.AccessToken = *parameter.Value
		case keyRefreshToken:
			sp.RefreshToken = *parameter.Value
		case keyExpiryTime:
			parsed, err := strconv.ParseInt(*parameter.Value, 10, 64)
			if err != nil {
				return StravaParams{}, err
			}
			sp.ExpiryTime = parsed
		}
	}
	return sp, nil
}

func (c *paramsClient) SetRefreshedParams(ctx context.Context, refreshToken, accessToken string, expiryTime int64) error {
	err := c.SetRefreshToken(ctx, refreshToken)
	if err != nil {
		return err
	}
	err = c.SetAccessToken(ctx, accessToken)
	if err != nil {
		return err
	}
	return c.SetExpiryTime(ctx, expiryTime)
}

func (c *paramsClient) SetRefreshToken(ctx context.Context, value string) error {
	return c.setParam(ctx, keyRefreshToken, value, types.ParameterTypeSecureString)
}

func (c *paramsClient) SetAccessToken(ctx context.Context, value string) error {
	return c.setParam(ctx, keyAccessToken, value, types.ParameterTypeSecureString)
}

func (c *paramsClient) SetExpiryTime(ctx context.Context, expiryTime int64) error {
	return c.setParam(ctx, keyExpiryTime, fmt.Sprintf("%d", expiryTime), types.ParameterTypeString)
}

func (c *paramsClient) setParam(ctx context.Context, key, value string, paramType types.ParameterType) error {
	_, err := c.ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      jsii.String(fmt.Sprintf("%s%s", prefix, key)),
		Value:     jsii.String(value),
		Type:      paramType,
		Overwrite: jsii.Bool(true),
	})
	return err
}

type StravaParams struct {
	ClientId     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	ExpiryTime   int64
}
