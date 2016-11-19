package spec

import (
	"fmt"
	"os"
)

type Spec struct {
	Name string                 `yaml:"name"`
	Type ItemType               `yaml:"type"`
	Spec map[string]interface{} `yaml:"spec"`
}

type ItemType string

var (
	Post ItemType = "Post"
)

func Load(specPath string) (*Spec, error) {
	if specPath == "" {
		return nil, fmt.Errorf("spec path not specified")
	}

	fp, err := os.Open(specPath)
	if err != nil {
		return nil, err
	}

	defer fp.Close()
	spec, err := Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", specPath, err)
	}

	return spec, nil
}