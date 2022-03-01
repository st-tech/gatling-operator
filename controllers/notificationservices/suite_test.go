package notificationservices

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBucket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NotificationServices Suite")
}
