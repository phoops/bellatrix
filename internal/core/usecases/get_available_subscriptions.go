package usecases

import (
	"github.com/phoops/ngsiv2/client"
	"github.com/phoops/ngsiv2/model"
	"github.com/pkg/errors"
)

type GetAvailableSubscriptions struct {
	orionClient *client.NgsiV2Client
}

func (u *GetAvailableSubscriptions) Execute(
	fiwareService string,
	servicePath string,
) ([]*model.Subscription, error) {
	response, err := u.orionClient.RetrieveSubscriptions(
		client.RetrieveSubscriptionsSetFiwareServicePath(servicePath),
		client.RetrieveSubscriptionsSetFiwareService(fiwareService),
	)

	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve subscriptions from context broker")
	}

	if len(response.Subscriptions) == 0 {
		return nil, nil
	}

	return response.Subscriptions, nil
}

// NewGetAvailableSubscriptions returns a new configured GetAvailableSubscriptions
// usecases
func NewGetAvailableSubscriptions(
	orionClient *client.NgsiV2Client,
) *GetAvailableSubscriptions {
	return &GetAvailableSubscriptions{
		orionClient: orionClient,
	}
}
