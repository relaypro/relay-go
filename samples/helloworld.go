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