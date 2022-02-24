Relay-go SDK

Use `go get github.com/relaypro/relay-go` to install

To create a workflow, create a func that accepts a `RelayApi` interface, and then add that workflow function with a name by calling `sdk.AddWorkflow(name, workflow)`

The interface `RelayApi` defines the event callbacks and requests that can be made. 

Pass callback functions to the `OnXXX()` functions, and in those functions make requests such as `api.Say()`

See cmd/main.go for examples