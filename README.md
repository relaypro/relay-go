# relay-go SDK

A Golang SDK for [Relay Workflows](https://developer.relaypro.com).

## Guides Documentation

The higher-level guides are available at https://developer.relaypro.com/docs


## API Reference Documentation

The generated documentation is available at https://relaypro.github.io/relay-go/

## Usage

- The following demonstrates a simple Hello World program, located in samples/helloworld.go
<pre>
// Copyright © 2022 Relay Inc.

package main

import (
    "relay-go/pkg/sdk"
    "fmt"
)

var port = ":8080"

func main() {

    sdk.AddWorkflow("helloworld", func(api sdk.RelayApi) {
        var sourceUri string
        
        api.OnStart(func(startEvent sdk.StartEvent) {
            sourceDevice := startEvent.Trigger.Args.SourceUri
            fmt.Println("starting interaction on source device", sourceDevice)
            api.StartInteraction(sourceDevice, "hello world")
        })
        
        api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
            if interactionLifecycleEvent.LifecycleType == "started" {
                sourceUri = interactionLifecycleEvent.SourceUri     // save the interaction id here to use in the timer callback
                var deviceName = api.GetDeviceName(sourceUri, false)
                api.SayAndWait(sourceUri, "What is your name?", "en-US")
                var pharses = []string {}
                var name = api.Listen(sourceUri, pharses, false, "en-US", 30)
                api.Say(sourceUri, "Hello " + name + " you are currently using " + deviceName, "en-US")
                api.EndInteraction(sourceUri, "hello world")
            }

            if interactionLifecycleEvent.LifecycleType == "ended" {
                fmt.Println("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
                api.Terminate()
            }
        })
    })
    
    sdk.InitializeRelaySdk(port)
}
</pre>

## Development

    bash 
    go get github.com/relaypro/relay-go
    cd relay-go

Start the demo workflow server:

    bash
    go run helloworld.go

To create a workflow, create a func that accepts a `RelayApi` interface, and then add that workflow function with a name by calling `sdk.AddWorkflow(name, workflow)`

The interface `RelayApi` defines the event callbacks and requests that can be made. 

Pass callback functions to the `OnXXX()` functions, and in those functions make requests such as `api.Say()`

## Additional Instructions for Development on Heroku

See the [Guide](https://developer.relaypro.com/docs/heroku).

## TLS Capability

Your workflow server must be exposed to the Relay server with TLS so
that the `wss` (WebSocket Secure) protocol is used, this is enforced by
the Relay server when registering workflows. See the
[Guide](https://developer.relaypro.com/docs/requirements) on this topic.


## License
[MIT](https://choosealicense.com/licenses/mit/)