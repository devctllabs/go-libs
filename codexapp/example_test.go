package codexapp_test

import (
	"context"
	"log"

	"github.com/devctllabs/go-libs/codexapp"
)

func ExampleOpen() {
	ctx := context.Background()
	client, err := codexapp.Open(ctx, codexapp.Config{
		ClientInfo: codexapp.ClientInfo{Name: "my-service", Version: "1.0.0"},
	})
	if err != nil {
		log.Print(err)
		return
	}
	defer func() { _ = client.Close(context.Background()) }()

	thread, err := client.StartThread(ctx, codexapp.StartThreadRequest{})
	if err != nil {
		log.Print(err)
		return
	}
	handle, err := client.StartTurn(ctx, codexapp.StartTurnRequest{
		ThreadID: thread.ID,
		Input:    []codexapp.Input{codexapp.Text("Review this repository")},
	})
	if err != nil {
		log.Print(err)
		return
	}
	_, _ = handle.Wait(ctx)
}
