package check

type HTTPCheckDetails struct {
	Details struct {
		AuthPassword    string              `json:"auth_password"`
		AuthUser        string              `json:"auth_user"`
		Body            string              `json:"body"`
		BodyMatches     []map[string]string `json:"body_matches"`
		FollowRedirects bool                `json:"follow_redirects"`
		Headers         map[string]string   `json:"headers"`
		IncludeBody     bool                `json:"include_body"`
		Method          string              `json:"method"`
		Url             string              `json:"url"`
	}
}

type HTTPCheckOut struct {
	CheckHeader
	HTTPCheckDetails
}
