package broker

import pubsubpb "github.com/iamBelugaa/grpc-pubsub/internal/generated/__proto__"

// ToString converts a pubsubpb.ResponseStatus enum to its string representation.
func ToString(s pubsubpb.ResponseStatus) string {
	switch s {
	case pubsubpb.ResponseStatus_ERROR:
		return "ERROR"
	case pubsubpb.ResponseStatus_OK:
		return "OK"
	default:
		return "UNSPECIFIED"
	}
}
