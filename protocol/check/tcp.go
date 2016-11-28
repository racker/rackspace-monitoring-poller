package check

type TCPCheckDetails struct {
	Details struct {
		BannerMatch string `json:"banner_match"`
		BodyMatch   string `json:"body_match"`
		Port        uint64 `json:"port"`
		SendBody    string `json:"send_body"`
		UseSSL      bool   `json:"ssl"`
	}
}

type TCPCheckOut struct {
	CheckHeader
	TCPCheckDetails
}
