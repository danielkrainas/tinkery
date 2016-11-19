package post

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/danielkrainas/tinkersnest/api/client"
	"github.com/danielkrainas/tinkersnest/api/v1"
	"github.com/danielkrainas/tinkersnest/cmd"
	"github.com/danielkrainas/tinkersnest/tinkerctl/spec"
)

func init() {
	cmd.Register("create", Info)
}

func run(ctx context.Context, args []string) error {
	specPath, ok := ctx.Value("flags.file").(string)
	if !ok {
		return errors.New("an object spec file path is required")
	}

	obj, err := spec.Load(specPath)
	if err != nil {
		return err
	}

	const ENDPOINT = "http://localhost:9240"

	c := client.New(ENDPOINT, http.DefaultClient)

	switch obj.Type {
	case spec.Post:
		post, err := postFromSpec(obj.Name, obj.Spec)
		if err != nil {
			return err
		}

		if post, err = c.Blog().CreatePost(post); err != nil {
			return err
		}

		fmt.Printf("post %q was created!\n", obj.Name)

	default:
		return fmt.Errorf("object type %q unsupported", obj.Type)
	}

	return nil
}

var (
	Info = &cmd.Info{
		Use:   "create",
		Short: "create an object on the server",
		Long:  "create an object on the server",
		Run:   cmd.ExecutorFunc(run),
		Flags: []*cmd.Flag{
			{
				Short:       "f",
				Long:        "file",
				Description: "object spec file",
				Type:        cmd.FlagString,
			},
		},
	}
)

func postFromSpec(name string, spec map[string]interface{}) (*v1.Post, error) {
	m, ok := spec["post"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing 'post' data in spec")
	}

	p := &v1.Post{
		Name:    name,
		Title:   m["title"].(string),
		Created: m["created"].(int64),
		Publish: false,
		Content: make([]*v1.Content, 0),
	}

	if publish, ok := m["publish"].(bool); ok {
		p.Publish = publish
	}

	contents, ok := m["content"].([]interface{})
	if !ok {
		return nil, errors.New("missing 'content' in post spec")
	}

	for _, c := range contents {
		if cm, ok := c.(map[string]interface{}); ok {
			c, err := getContent(cm)
			if err != nil {
				return nil, err
			}

			p.Content = append(p.Content, c)
		}
	}

	return p, nil
}

func getContent(spec map[string]interface{}) (*v1.Content, error) {
	c := &v1.Content{}
	if t, ok := spec["type"].(string); !ok {
		return nil, errors.New("invalid or missing content 'type' in spec")
	} else {
		c.Type = t
	}

	if sdata, ok := spec["data"].(string); ok {
		c.Data = []byte(sdata)
	} else if src, ok := spec["src"].(string); ok {
		data, err := ioutil.ReadFile(src)
		if err != nil {
			return nil, err
		}

		c.Data = data
	}

	if c.Data == nil {
		return nil, errors.New("content does not have any data associated")
	}

	return c, nil
}