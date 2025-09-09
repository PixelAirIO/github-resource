package githubresource

type BaseRequest struct {
	Source struct {
		Kind   string `json:"kind"`
		Config Config `json:"config"`
	} `json:"source"`
}

type Kind interface {
	Check(stdin []byte)
	In(stdin []byte, dest string)
	Out(stdin []byte, src string)
}
