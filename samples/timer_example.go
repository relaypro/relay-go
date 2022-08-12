// Copyright © 2022 Relay Inc.

package main

import (
    "relay-go/pkg/sdk"   
    "fmt"
)

func timer_example(api sdk.RelayApi) {

    var sourceUri string

    api.OnStart(func(startEvent sdk.StartEvent) {
        sourceDevice := startEvent.Trigger.Args.SourceUri
        fmt.Println("starting interaction on source device", sourceDevice)
        api.StartInteraction(sourceDevice, "hello")
    })
    
    api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
        if interactionLifecycleEvent.LifecycleType == "started" {
            fmt.Println("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
            sourceUri = interactionLifecycleEvent.SourceUri     // save the interaction id here to use in the timer callback
            api.Say(sourceUri, "Hello world", "")
            api.EndInteraction(sourceUri, "hello")
        }

        if interactionLifecycleEvent.LifecycleType == "ended" {
            fmt.Println("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
            api.Terminate()
        }
    })
        
    
}
