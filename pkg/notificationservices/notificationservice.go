package notificationservices

import (
	"sync"
)

type NotificationServiceProvider interface {
	GetName() string
	Notify(gatlingName string, reportURL string, secretData map[string][]byte) error
}

// use sync.Map to achieve thread safe read and write to map
var notificationServiceProvidersSyncMap = &sync.Map{}

func GetProvider(provider string) *NotificationServiceProvider {
	v, ok := notificationServiceProvidersSyncMap.Load(provider)
	if !ok {
		var nsp NotificationServiceProvider
		switch provider {
		case "slack":
			nsp = &SlackNotificationServiceProvider{providerName: provider}
		default:
			return nil
		}
		v, _ = notificationServiceProvidersSyncMap.LoadOrStore(provider, &nsp)
	}
	return v.(*NotificationServiceProvider)
}
