package usecases

import (
	"github.com/phoops/bellatrix/internal/core/entities"
	"github.com/phoops/ngsiv2/client"
	"github.com/phoops/ngsiv2/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ApplySubscriptionsPatches struct {
	orionClient *client.NgsiV2Client
	logger      *zap.Logger
}

func NewApplySubscriptionsPatches(orionClient *client.NgsiV2Client, logger *zap.Logger) *ApplySubscriptionsPatches {
	return &ApplySubscriptionsPatches{orionClient: orionClient, logger: logger}
}

func (u *ApplySubscriptionsPatches) Execute(
	subscriptionsPatches []*entities.SubscriptionsPatch,
) error {
	if len(subscriptionsPatches) == 0 {
		u.logger.Info(
			"Subscriptions state in sync. No changes needed.",
		)

		return nil
	}
	// apply the patches to the service broker
	for _, patch := range subscriptionsPatches {
		err := u.applyAddSubscriptionsPatch(
			patch.SubscriptionsToAdd,
			patch.FiwareService,
			patch.ServicePath,
		)

		if err != nil {
			return err
		}

		err = u.applyDeleteSubscriptionsPatch(
			patch.SubscriptionsToDelete,
			patch.FiwareService,
			patch.ServicePath,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func (u *ApplySubscriptionsPatches) applyAddSubscriptionsPatch(
	subs []*model.Subscription,
	fiwareService string,
	fiwareServicePath string,
) error {
	for _, sub := range subs {
		u.logger.Info(
			"Add patch, adding subscription",
			zap.String("subscription_description", sub.Description),
		)
		_, err := u.orionClient.CreateSubscription(
			sub,
			client.SubscriptionSetFiwareService(fiwareService),
			client.SubscriptionSetFiwareServicePath(fiwareServicePath),
		)

		if err != nil {
			return errors.Wrapf(
				err,
				"could not apply the add subscription patch for subscription with description %s",
				sub.Description,
			)
		}
	}
	return nil
}

func (u *ApplySubscriptionsPatches) applyDeleteSubscriptionsPatch(
	subs []*model.Subscription,
	fiwareService string,
	fiwareServicePath string,
) error {
	for _, sub := range subs {
		u.logger.Info(
			"Delete patch, deleting subscription",
			zap.String("subscription_id", sub.Id),
			zap.String("subscription_description", sub.Description),
		)
		err := u.orionClient.DeleteSubscription(
			sub.Id,
			client.SubscriptionSetFiwareService(fiwareService),
			client.SubscriptionSetFiwareServicePath(fiwareServicePath),
		)

		if err != nil {
			return errors.Wrapf(
				err,
				"could not apply the delete subscription patch for subscription with description %s",
				sub.Description,
			)
		}
	}
	return nil
}
