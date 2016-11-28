package check

type PingCheckDetails struct {
	Details struct {
		Count uint8 `json:"count"`
	}
}

type PingCheckOut struct {
	CheckHeader
	PingCheckDetails
}
