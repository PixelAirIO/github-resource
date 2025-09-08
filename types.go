package githubresource

type BaseRequest struct {
	Source struct {
		Kind   string `json:"kind"`
		Config Config `json:"config"`
	} `json:"source"`
}
