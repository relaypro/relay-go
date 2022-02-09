package main

import (
    // relay-go     -> module name
    // pkg          -> local file path
    // sdk          -> package name
    // now can use anything exported from any file in the sdk package with sdk.Whatever()
    "relay-go/pkg/sdk"
    
    "fmt"
)



func main() {
    fmt.Println("hello world")
        
    sdk.AddWorkflow("hello", func(api sdk.RelayApi) {
        api.OnStart(func(startEvent sdk.StartEvent) {
            fmt.Println("i'm a callback for START events", startEvent)
            
            source := startEvent.Trigger.Args.SourceUri
            fmt.Println("source uri", source)
            
            api.StartInteraction(source)
        })
        
        api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
            fmt.Println("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
            
            api.Say(interactionLifecycleEvent.SourceUri, "hello world")
            
            api.Terminate()
        })
    })
    
    sdk.InitializeRelaySdk()
}

// create the http server listening on chosen port

// pass the server to the initializeRelaySdk function

// initializeRelaySdk creates a websocket listener from the http server, sets up event listeners on the ws, and returns an interface that allows you to set workflow modules with a name



//createWorkflow is used by the wf module to create a Workflow interface, which takes a RelayEventAdapter in ctor, which is a class that takes a websocket

// createworkflow takes a function and then calls it, passing in the relay event adapter, which was instantiated with the websocket connection, and provides all the on event callbacks, and say, etc functions
/*export default createWorkflow(relay: RelayEventAdapter => {

  relay.on(Event.START, async () => {
    // ...
  }
*/

