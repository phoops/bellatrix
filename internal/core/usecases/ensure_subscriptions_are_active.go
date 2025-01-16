package usecases

import (
	"github.com/phoops/bellatrix/internal/core/entities"
	"github.com/phoops/ngsiv2/client"
	"github.com/phoops/ngsiv2/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type EnsureSubscriptionsAreActive struct {
	getAvailableSubscriptions *GetAvailableSubscriptions
	orionClient               *client.NgsiV2Client
	logger                    *zap.SugaredLogger
	instancePrefix            string
}

func NewEnsureSubscriptionsAreActive(
	getAvailableSubscriptions *GetAvailableSubscriptions,
	logger *zap.SugaredLogger,
	client *client.NgsiV2Client,
	instancePrefix string,
) *EnsureSubscriptionsAreActive {
	return &EnsureSubscriptionsAreActive{instancePrefix: instancePrefix, getAvailableSubscriptions: getAvailableSubscriptions, logger: logger, orionClient: client}
}

func (u *EnsureSubscriptionsAreActive) Execute(
	requestedSubscriptions []entities.SubscriptionRequest,
) error {
	for _, request := range requestedSubscriptions {
		subscriptionsInOrion, err := u.getAvailableSubscriptions.Execute(
			request.FiwareService,
			request.ServicePath,
		)

		orionSubsManagedByBellatrix := getSubscriptionsManagedByBellatrix(subscriptionsInOrion, u.instancePrefix)

		if err != nil {
			return errors.Wrapf(
				err,
				"could not get subscriptions on context broker for servicePath %s, and fiwareService %s, during ensure subscriptions are active",
				request.ServicePath,
				request.FiwareService,
			)
		}

		for _, subsForServicePath := range orionSubsManagedByBellatrix {
			if isSubscriptionFailed(subsForServicePath) {
				u.logger.Warnw(
					"Subscription is in failed state, need to recreate.",
					"subscription_id",
					subsForServicePath.Id,
					"failure_date",
					subsForServicePath.Notification.LastFailure,
				)

				// delete subscription than recreate

				err := u.orionClient.DeleteSubscription(
					subsForServicePath.Id,
					client.SubscriptionSetFiwareService(request.FiwareService),
					client.SubscriptionSetFiwareServicePath(request.ServicePath),
				)

				if err != nil {
					return errors.Wrapf(
						err,
						"could not delete failed subscriptions with id %s - name: %s",
						subsForServicePath.Id,
						subsForServicePath.Description,
					)
				}

				u.logger.Infow(
					"Deleted failed subscription",
					"subscription_id",
					subsForServicePath.Id,
					"failure_date",
					subsForServicePath.Notification.LastFailure,
					"last_success_code",
					subsForServicePath.Notification.LastSuccessCode,
				)

				subInState, err := findSubscriptionInsideSubState(
					request.Subscriptions,
					subsForServicePath.Description,
				)

				if err != nil {
					return errors.Wrapf(
						err,
						"could not found subscription to recreate in state - name: %s",
						subsForServicePath.Description,
					)
				}

				newSubscription := &model.Subscription{
					Description:  subsForServicePath.Description,
					Subject:      subInState.Subject,
					Notification: subInState.Notification,
				}

				_, err = u.orionClient.CreateSubscription(
					newSubscription,
					client.SubscriptionSetFiwareService(request.FiwareService),
					client.SubscriptionSetFiwareServicePath(request.ServicePath),
				)

				if err != nil {
					return errors.Wrapf(
						err,
						"could not recreate failed subscription with id %s - name: %s",
						subsForServicePath.Id,
						subsForServicePath.Description,
					)
				}

				u.logger.Infow(
					"Recreated failed subscription",
					"name",
					subsForServicePath.Description,
				)
			}

		}

	}
	return nil
}

// We need to check for last failure date, and for the last success code
// because @telefonica 404 MEANS SUCCESS
func isSubscriptionFailed(subscription *model.Subscription) bool {
	return (subscription.Notification != nil && subscription.Notification.LastFailure != nil) ||
		(subscription.Notification != nil && subscription.Notification.LastSuccessCode != nil && *subscription.Notification.LastSuccessCode >= 300)
}

func findSubscriptionInsideSubState(
	subscriptionsInState []*model.Subscription,
	subscriptionDescription string,
) (*model.Subscription, error) {
	for _, subInState := range subscriptionsInState {
		if subInState.Description == subscriptionDescription {
			return subInState, nil
		}
	}
	return nil, errors.New("could not find the subscription to create inside subscriptions state")
}
