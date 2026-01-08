package main

import (
	"os"

	"github.com/blue-monads/turnix/backend/services/corehub/buddyhub/funnel"
	"github.com/blue-monads/turnix/backend/utils/qq"
)

func main() {

	qq.Println("@main/1")

	// http://test.localhost:8080

	client := funnel.NewFunnelClient(funnel.FunnelClientOptions{
		LocalHttpPort:   8000,
		RemoteFunnelUrl: "http://test.localhost:8080/funnel/register/test",
		ServerId:        "test",
	})
	defer client.Stop()

	qq.Println("@main/2")

	err := client.Start("")
	if err != nil {
		qq.Println("Error starting client: %v", err)
		os.Exit(1)
	}

	qq.Println("@main/3")
}
