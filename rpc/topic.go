package rpc

import "fmt"

const (
	// topicPrefix is the root prefix for all courier MQTT topics.
	topicPrefix = "mrpc"
)

// RequestTopic returns the topic that clients publish to when calling serviceName.
// Example: "mrpc/request/UserService"
func RequestTopic(serviceName string) string {
	return fmt.Sprintf("%s/request/%s", topicPrefix, serviceName)
}

// SharedRequestTopic returns the shared subscription topic for serviceName.
// Example: "$share/UserService/mrpc/request/UserService"
//
// All server instances providing the same service subscribe to this topic.
// The broker delivers each message to exactly one instance in the group.
func SharedRequestTopic(serviceName string) string {
	return fmt.Sprintf("$share/%s/%s/request/%s", serviceName, topicPrefix, serviceName)
}

// ResponseTopic returns the topic where responses are sent for a specific device.
// Example: "mrpc/response/device-abc123"
func ResponseTopic(deviceID string) string {
	return fmt.Sprintf("%s/response/%s", topicPrefix, deviceID)
}

// EventTopic returns the topic for service-to-client event publishing.
// Example: "mrpc/event/UserService/UserOnline"
func EventTopic(serviceName, eventName string) string {
	return fmt.Sprintf("%s/event/%s/%s", topicPrefix, serviceName, eventName)
}
