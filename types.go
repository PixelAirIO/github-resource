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

type Metadata []*MetadataField

func (m *Metadata) Add(name, value string) {
	*m = append(*m, &MetadataField{Name: name, Value: value})
}

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
