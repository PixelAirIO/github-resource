package githubresource

import (
	"os"
)

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

func InterpolateBuildMetadata(s string) string {
	return os.Expand(s, func(e string) string {
		switch e {
		case "BUILD_ID",
			"BUILD_NAME",
			"BUILD_JOB_NAME",
			"BUILD_PIPELINE_NAME",
			"BUILD_PIPELINE_INSTANCE_VARS",
			"BUILD_CREATED_BY",
			"BUILD_TEAM_NAME",
			"ATC_EXTERNAL_URL",
			"BUILD_URL",
			"BUILD_URL_SHORT":
			return os.Getenv(e)
		}
		return "$" + e
	})
}
