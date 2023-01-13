
// Copyright © 2022 Relay Inc.

package main

import (
    "relay-go/pkg/sdk"
    log "github.com/sirupsen/logrus"
)

var port = ":8080"

func main() {
    log.SetLevel(log.InfoLevel)

    sdk.AddWorkflow("hellopath", func(api sdk.RelayApi) {
        var sourceUri string
        
        api.OnStart(func(startEvent sdk.StartEvent) {
            sourceUri := api.GetSourceUri(startEvent)
            log.Debug("Started hello wf from sourceUri: ", sourceUri, " trigger: ", startEvent.Trigger)
            api.StartInteraction(sourceUri, "hello interaction")
        })
        
        api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
            log.Debug("User workflow got interaction lifecycle: ", interactionLifecycleEvent)

            if interactionLifecycleEvent.LifecycleType == "started" {
                sourceUri = interactionLifecycleEvent.SourceUri
                var deviceName = api.GetDeviceName(sourceUri, false)
                api.SayAndWait(sourceUri, "What is your name?", sdk.ENGLISH)
                var name = api.Listen(sourceUri, []string {}, false, sdk.ENGLISH, 30)
                api.Say(sourceUri, "Hello " + name + " you are currently using " + deviceName, sdk.ENGLISH)
                api.EndInteraction(sourceUri)
            }

            if interactionLifecycleEvent.LifecycleType == "ended" {
                log.Debug("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
                api.Terminate()
            }
        })
    })
    
    sdk.InitializeRelaySdk(port)
}