package entities

import "github.com/phoops/ngsiv2/model"

// SubscriptionRequest represent a request for a subscription
// bellatrix will try to satisfy the requested state for the subscription
type SubscriptionRequest struct {
	ServicePath   string                `json:"service_path,omitempty"`
	FiwareService string                `json:"fiware_service,omitempty"`
	Subscriptions []*model.Subscription `json:"subscriptions"`
}

// OrionClientOptions represent options for the main orion client
// you can specify only one context broker in which you run all the
// subscriptions requests
type OrionClientOptions struct {
	AdditionalHeaders map[string]string `json:"additional_headers,omitempty"`
	ClientURL         string            `json:"client_url"`
}

// SubscriptionsRequestedState represent the main state you can request
// in order to have the subscriptions synced with the context broker
type SubscriptionsRequestedState struct {
	ClientOptions      OrionClientOptions    `json:"client_options"`
	SubscriptionsState []SubscriptionRequest `json:"subscriptions_state"`
}

type SubscriptionsPatch struct {
	ServicePath           string                `json:"service_path,omitempty"`
	FiwareService         string                `json:"fiware_service,omitempty"`
	SubscriptionsToAdd    []*model.Subscription `json:"subscriptions_to_add"`
	SubscriptionsToDelete []*model.Subscription `json:"subscriptions_to_delete"`
}
