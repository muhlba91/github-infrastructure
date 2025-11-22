package github

// HasSubscription checks if a subscription is present.
// subscription: the subscription to check.
func HasSubscription(subscription *string) bool {
	return subscription != nil && *subscription != "none"
}
