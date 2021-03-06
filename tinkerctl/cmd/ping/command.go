package ping

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielkrainas/gobag/cmd"

	"github.com/danielkrainas/tinkersnest/api/client"
)

func init() {
	cmd.Register("ping", Info)
}

func run(ctx context.Context, args []string) error {
	const ENDPOINT = "http://localhost:9240"

	c := client.New(ENDPOINT, http.DefaultClient)

	err := c.Ping()
	if err != nil {
		fmt.Printf("error sending ping: %v\n", err)
		return nil
	}

	fmt.Println("Ok.")
	return nil
}

var (
	Info = &cmd.Info{
		Use:   "ping",
		Short: "`ping`",
		Long:  "`ping`",
		Run:   cmd.ExecutorFunc(run),
	}
)
