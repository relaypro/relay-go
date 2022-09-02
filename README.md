# relay-go SDK

A Golang SDK for [Relay Workflows](https://developer.relaypro.com).

## Guides Documentation

The higher-level guides are available at https://developer.relaypro.com/docs


## API Reference Documentation

The generated documentation is available at https://relaypro.github.io/relay-go/

## Usage

- The following demonstrates a simple Hello World program, located in samples/helloworld.go.
It also uses [github.com/gorilla/websocket](https://github.com/gorilla/websocket) as a 
websocket implementation.
<pre>
// Copyright © 2022 Relay Inc.

package main

import (
    "relay-go/pkg/sdk"
    log "github.com/sirupsen/logrus"
)

var port = ":8080"

func main() {
    log.SetLevel(log.DebugLevel)

    sdk.AddWorkflow("helloworld", func(api sdk.RelayApi) {
        var sourceUri string
        
        api.OnStart(func(startEvent sdk.StartEvent) {
            sourceUri := startEvent.Trigger.Args.SourceUri
            log.Debug("Started hello wf from sourceUri: ", sourceUri, " trigger: ", startEvent.Trigger)
            api.StartInteraction(sourceUri, "hello interaction")
        })
        
        api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
            log.Debug("User workflow got interaction lifecycle: ", interactionLifecycleEvent)

            if interactionLifecycleEvent.LifecycleType == "started" {
                sourceUri = interactionLifecycleEvent.SourceUri     // save the interaction id here to use in the timer callback
                var deviceName = api.GetDeviceName(sourceUri, false)
                api.SayAndWait(sourceUri, "What is your name?", "en-US")
                var pharses = []string {}
                var name = api.Listen(sourceUri, pharses, false, "en-US", 30)
                api.Say(sourceUri, "Hello " + name + " you are currently using " + deviceName, "en-US")
                api.EndInteraction(sourceUri, "hello interaction")
            }

            if interactionLifecycleEvent.LifecycleType == "ended" {
                log.Debug("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
                api.Terminate()
            }
        })
    })
    
    sdk.InitializeRelaySdk(port)
}
</pre>

# Running the Sample

Use git to make a local copy of the SDK and include sample application:

    git clone git@github.com:relaypro/relay-go.git
    cd relay-go/samples
    go run helloworld.go

However, until you register a trigger with the server via the Relay CLI, the
workflow won't get invoked via an inbound websocket connection.
[Run a built-in workflow](https://developer.relaypro.com/docs/run-a-built-in-workflow)
first to verify your Relay CLI installation. Then follow the
[documentation for a custom workflow](https://developer.relaypro.com/docs/getting-started).

The code is split into two subdirectories: `pkg/sdk` and `samples`.

The `pkg/sdk` contains Relay SDK functionality, `app` contains sample code for creating and 
running workflows using the SDK.  The `app` includes the sample workflow.

The main entry point is the `relay-go/pkg/sdk` package.  To create a workflow, add the import 
`relay-go/pkg/sdk` and then add that workflow with a URL path name by calling 
`sdk.AddWorkflow("helloworld", func(api sdk.RelayApi)`.  See realy-go/samples for an example implementation.

The utility `api.go` defines the event allbacks that your workflow can respond to, as well as 
the definitions to the requests that can be made inside those callbacks.

The websocket server implementation is provided in the `ws_receiver.go` file to receive websocket
client connections and start workflows when requested by the Relay server.

Now, you can start the demo workflow server:

    go run helloworld.go

## TLS Capability

Your workflow server must be exposed to the Relay server with TLS so
that the `wss` (WebSocket Secure) protocol is used, this is enforced by
the Relay server when registering workflows. See the
[Guide](https://developer.relaypro.com/docs/requirements) on this topic.

## Verbose Mode Logging

The SDK is using [Logrus](https://github.com/sirupsen/logrus) for logging.  Logging levels can
be configured at the top of your workflow file using Logrus's `SetLevel` function.  The sample
applications will show log messages from INFO level and above.  If you wish to see more logging
detail from the SDK, and you continue to use Logrus, then call the `SetLevel` function to use
the DEBUG log level:

    log.SetLevel(DebugLevel)

## License
[MIT](https://choosealicense.com/licenses/mit/)