package factory

import (
	"encoding/json"
	"log"
	"strings"

	ghr "github.com/PixelAirIO/github-resource"
	cfp "github.com/PixelAirIO/github-resource/kinds/commits_from_prs"
	"github.com/PixelAirIO/github-resource/kinds/pr"
	"github.com/PixelAirIO/github-resource/kinds/prs"
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
	case "pr":
		return &pr.Pr{}
	case "commits-from-prs":
		return &cfp.CommitsFromPrs{}
	default:
		log.Fatalf("unknown kind: %s", req.Source.Kind)
	}

	return nil
}
