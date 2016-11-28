package check

import (
	"encoding/json"
	"log"
)

// Given a received check request, this will unmarshal the request into one of the known polymorphic types.
// This method needs to be updated to add to the known types.
func NewCheck(rawParams json.RawMessage) Check {
	checkBase := &CheckBase{}

	err := json.Unmarshal(rawParams, &checkBase)
	if err != nil {
		log.Printf("Error unmarshalling checkbase")
		return nil
	}
	switch checkBase.CheckType {
	case "remote.tcp":
		return NewTCPCheck(checkBase)
	case "remote.http":
		return NewHTTPCheck(checkBase)
	case "remote.ping":
		return NewPingCheck(checkBase)
	default:

		log.Printf("Invalid check type: %v", checkBase.CheckType)
	}
	return nil
}
