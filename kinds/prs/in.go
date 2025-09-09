package prs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type inRequest struct {
	Source  Source  `json:"source"`
	Version version `json:"version"`
}

type inResponse struct {
	Version version `json:"version"`
}

func (*Prs) In(stdin []byte, dest string) {
	dc := json.NewDecoder(bytes.NewReader(stdin))
	dc.DisallowUnknownFields()

	request := &inRequest{}
	err := dc.Decode(request)
	if err != nil {
		log.Fatalf("failed to unmarshal in request: %v", err)
	}

	if request.Version.Prs == "" {
		log.Fatal("empty list of Pr's passed in")
	}

	prs := strings.Split(request.Version.Prs, ",")
	if len(prs) == 0 {
		log.Fatal("got an empty list of Pr's after trying to parse the version")
	}

	prsMsh, err := json.Marshal(prs)
	if err != nil {
		log.Fatalf("error marshaling PR's to write to disk: %v", err)
	}

	err = os.WriteFile(filepath.Join(dest, "prs.json"), prsMsh, 0644)
	if err != nil {
		log.Fatalf("error writing prs.json: %v", err)
	}

	resp := inResponse{
		Version: request.Version,
	}

	ver, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshaling: %v", err)
	}

	fmt.Println(string(ver))
}
