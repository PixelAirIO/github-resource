package githubresource

type BaseRequest struct {
	Source `json:"source"`
}

type Source struct {
	Config
	Kind string `json:"kind"`
}

type Kind interface {
	Check(stdin []byte)
	In(stdin []byte, dest string)
	Out(stdin []byte, src string)
}
