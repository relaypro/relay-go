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
        api.StartInteraction(sourceDevice)
    })
    
    api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
        if interactionLifecycleEvent.LifecycleType == "started" {
            sourceUri = interactionLifecycleEvent.SourceUri     // save the interaction id here to use in the timer callback
            
            api.Say(interactionLifecycleEvent.SourceUri, "starting timer", "")

            // start timer
            api.SetTimer(sdk.TIMEOUT_TIMER_TYPE, "timer 1", 5, sdk.SECS_TIMEOUT_TYPE)
        }
    })
    
    api.OnTimerFired(func(timerFiredEvent sdk.TimerFiredEvent) {
        fmt.Println("timer fired! name ", timerFiredEvent.Name)
        if timerFiredEvent.Name == "timer 1" {
            api.Say(sourceUri, "timer fired", "")
            api.Terminate()
        }
    })
}
