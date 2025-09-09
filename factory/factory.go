package factory

import (
	"encoding/json"
	"log"
	"strings"

	ghr "github.com/pixel-air/github-resource"
	"github.com/pixel-air/github-resource/kinds/prs"
)

func NewKind(stdin []byte) ghr.Kind {
	var req ghr.BaseRequest
	err := json.Unmarshal(stdin, &req)
	if err != nil {
		log.Fatalf("failed to unmarshal to base request: %v", err)
	}

	switch strings.ToLower(req.Source.Kind) {
	case "prs":
		return &prs.Prs{}
	default:
		log.Fatalf("unknown kind: %s", req.Source.Kind)
	}

	return nil
}
