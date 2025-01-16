package usecases

import (
	"strings"

	"go.uber.org/zap"

	"github.com/phoops/bellatrix/internal/core/entities"
	"github.com/phoops/ngsiv2/model"
	"github.com/pkg/errors"
)

// BellatrixManagedSubscriptionsPrefix represent the prefix of the subscription descriptions
// managed by bellatrix, so we don't consider other subscriptions, made by hand or for other purposes
// when we crud the subscriptions
const BellatrixManagedSubscriptionsPrefix = "BELLATRIX_MANAGED_"

type GetSubscriptionsPatches struct {
	getAvailableSubscriptions *GetAvailableSubscriptions
	logger                    *zap.Logger
	instancePrefix            string
}

func NewGetSubscriptionsPatches(
	getAvailableSubscriptions *GetAvailableSubscriptions,
	logger *zap.Logger,
	instancePrefix string,
) *GetSubscriptionsPatches {
	return &GetSubscriptionsPatches{getAvailableSubscriptions: getAvailableSubscriptions, logger: logger, instancePrefix: instancePrefix}
}

func (u *GetSubscriptionsPatches) Execute(
	requestedSubscriptions []entities.SubscriptionRequest,
) ([]*entities.SubscriptionsPatch, error) {

	var subsPatches []*entities.SubscriptionsPatch
	// for each subscription request, we will check the managed bellatrix subscriptions
	// on the context broker, for each service/servicepath specified in each request
	for _, request := range requestedSubscriptions {
		subscriptionsInOrion, err := u.getAvailableSubscriptions.Execute(
			request.FiwareService,
			request.ServicePath,
		)

		orionSubsManagedByBellatrix := getSubscriptionsManagedByBellatrix(subscriptionsInOrion, u.instancePrefix)

		if err != nil {
			return nil, errors.Wrapf(
				err,
				"could not get subscriptions on context broker for servicePath %s, and fiwareService %s",
				request.ServicePath,
				request.FiwareService,
			)
		}

		// we will check the desired subscriptions passed as parameter
		// against the subscriptions managed by bellatrix
		// and we will apply the add/delete patches in order to match the
		// desired state
		subscriptionsToAdd, subscriptionsToDelete := getBellatrixSubscriptionsDiff(
			request.Subscriptions,
			orionSubsManagedByBellatrix,
		)
		u.logger.Debug(
			"Subscriptions diff",
			zap.Any("subscriptions_to_delete", subscriptionsToDelete),
			zap.Any("subscriptions_to_add", subscriptionsToAdd),
			zap.String("fiware_service", request.FiwareService),
			zap.String("fiware_service_path", request.ServicePath),
		)

		if len(subscriptionsToAdd) != 0 || len(subscriptionsToDelete) != 0 {
			subsPatches = append(subsPatches, &entities.SubscriptionsPatch{
				ServicePath:           request.ServicePath,
				FiwareService:         request.FiwareService,
				SubscriptionsToAdd:    subscriptionsToAdd,
				SubscriptionsToDelete: subscriptionsToDelete,
			})
		}
	}
	return subsPatches, nil
}

func getSubscriptionsManagedByBellatrix(
	subscriptions []*model.Subscription,
	instancePrefix string,
) []*model.Subscription {
	var managedSubscriptions []*model.Subscription
	// full prefix is the join of instance prefix and bellatrix prefix
	fullPrefix := instancePrefix + BellatrixManagedSubscriptionsPrefix
	for _, sub := range subscriptions {
		if strings.HasPrefix(sub.Description, fullPrefix) {
			managedSubscriptions = append(managedSubscriptions, sub)
		}
	}
	return managedSubscriptions
}

// getBellatrixSubscriptionsDiff will check the differences between the subscriptions
// in orion and the desired subscriptions state
// the subscriptions involved in this comparison have the difference populated
// with the bellatrix prefix
// we assume this.
func getBellatrixSubscriptionsDiff(
	subscriptionDesiredState []*model.Subscription,
	subscriptionsInOrion []*model.Subscription,
) ([]*model.Subscription, []*model.Subscription) {
	subscriptionsToDelete := bellatrixSubscriptionsDifferenceByID(
		subscriptionsInOrion,
		subscriptionDesiredState,
	)
	subscriptionsToAdd := bellatrixSubscriptionsDifferenceByID(
		subscriptionDesiredState,
		subscriptionsInOrion,
	)
	return subscriptionsToAdd, subscriptionsToDelete
}

func bellatrixSubscriptionsDifferenceByID(
	subsA []*model.Subscription,
	subsB []*model.Subscription,
) []*model.Subscription {
	var diff []*model.Subscription

	subsMap := make(map[string]bool)

	for _, item := range subsB {
		subsMap[item.Description] = true
	}

	for _, item := range subsA {
		if _, ok := subsMap[item.Description]; !ok {
			diff = append(diff, item)
		}
	}
	return diff
}
