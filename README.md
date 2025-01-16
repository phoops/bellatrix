# Bellatrix - Orion subscription manager

Bellatrix will sync your orion subscription with a `state file`.

This program is stateless and idempotent.

It's like `ansible`, without jinja templates, inconsistency and design flaws, for `orion subscriptions`.

## State file

```json
{
  "client_options": {
    "client_url": "<context-broker-url>",
    "additional_headers": { // Bellatrix sends this headers on every request to context broker
      "Authorization": "Bearer you token",
      "X-CUSTOM-TOKEN": "custom-value",
    }
  },
  "subscriptionsState": [ // Array of subscriptions state
    {
      "service_path": "/REPLACE_WITH_ORION_SERVICE_PATH", // optional
      "fiware_service": "REPLACE_WITH_ORION_FIWARE", //optional
      "subscriptions": [ // Array of orion subscriptions requests, see the api reference.
        {
          "description": "REPLACE_WITH_THE_SUBSCRIPTION_DESCRIPTION", // this should be unique across all the subscriptions.
          "subject": {
            "entities": [
              {
                "idPattern": ".*",
                "type": "REPLACE_WITH_THE ENTITY_TYPE"
              }
            ]
          },
          "notification": {
            "httpCustom": {
              "url": "REPLACE_WITH_ORION_URL_FOR_NOTIFICATION",
              "headers": {
                "Authorization": "Basic REPLACE_WITH_THE_BASIC_AUTHENTICATION_VALUE"
              }
            }
          }
        }
      ]
    }
  ]
}

```

Bellatrix will sync the context broker, in order to match the subscriptions contained in this state file.

Bellatrix will not conisder subscriptions not managed by itself, so you can manually add subscriptions, and leave the bellatrix state unchanged.

In order to add/remove subscriptions, just remove the items from subscriptions array.

A side note for deletion, in order to delete properly all the subscriptions from a particular `fiware-service` or `service-path`, first remove the items from `subscriptions` array, apply bellatrix, so it will remove all the subscriptions from context broker then remove the item from `subscriptionsState` array, for the particular `fiware-service` or `service-broker` you are targeting


