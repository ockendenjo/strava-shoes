package bagging

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const pk = "ID"
const expiry = "Expiry"

type Client interface {
	HasId(ctx context.Context, d int64) (bool, error)
	PutId(ctx context.Context, id int64) error
}

func NewClient(dbClient *dynamodb.Client, tableName string) Client {
	return &baggingClient{dbClient: dbClient, tableName: tableName}
}

type baggingClient struct {
	dbClient  *dynamodb.Client
	tableName string
}

func (b baggingClient) HasId(ctx context.Context, id int64) (bool, error) {
	res, err := b.dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(b.tableName),
		Key: map[string]types.AttributeValue{
			pk: &types.AttributeValueMemberS{
				Value: fmt.Sprint(id),
			},
		},
	})
	if err != nil {
		return false, err
	}
	return res.Item != nil, nil
}

func (b baggingClient) PutId(ctx context.Context, id int64) error {

	expiryTime := time.Now().AddDate(0, 2, 0).Unix()

	_, err := b.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(b.tableName),
		Item: map[string]types.AttributeValue{
			pk: &types.AttributeValueMemberS{
				Value: fmt.Sprint(id),
			},
			expiry: &types.AttributeValueMemberN{
				Value: fmt.Sprint(expiryTime),
			},
		},
	})
	return err
}
