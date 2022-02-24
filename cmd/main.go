package main

import (
    "relay-go/pkg/sdk"
    "fmt"
)

var port = ":5000"

func main() {
    // add workflow functions by name
    sdk.AddWorkflow("timer", timer_example)
    
    sdk.AddWorkflow("hello", func(api sdk.RelayApi) {
        var sourceUri string
        
        api.OnStart(func(startEvent sdk.StartEvent) {
            sourceDevice := startEvent.Trigger.Args.SourceUri
            fmt.Println("starting interaction on source device", sourceDevice)
            api.StartInteraction(sourceDevice)
        })
        
        api.OnInteractionLifecycle(func(interactionLifecycleEvent sdk.InteractionLifecycleEvent) {
            if interactionLifecycleEvent.LifecycleType == "started" {
                fmt.Println("i'm a callback for interaction lifecycle: ", interactionLifecycleEvent)
                sourceUri = interactionLifecycleEvent.SourceUri     // save the interaction id here to use in the timer callback
                
                // say
                api.Say(interactionLifecycleEvent.SourceUri, "hello world", "")
                
                // play
//                 id := api.Play(interactionLifecycleEvent.SourceUri, "ibot-priv:///incoming_call_ring3.opus")
//                 fmt.Println("play id ", id)
                
                // leds
//                 api.Rainbow(interactionLifecycleEvent.SourceUri, -1)
//                 api.Flash(interactionLifecycleEvent.SourceUri, "FF0000")
//                 api.SwitchAllLedOn(interactionLifecycleEvent.SourceUri, "00FF00")
//                 api.Rotate(interactionLifecycleEvent.SourceUri, "00FF00")
//                 api.SwitchAllLedOn(interactionLifecycleEvent.SourceUri, "FF00FF")
//                 api.Breathe(interactionLifecycleEvent.SourceUri, "FFFF00")
                
                // vibrate
//                 api.Vibrate(interactionLifecycleEvent.SourceUri, []uint64{100, 500, 500,  500, 500, 500})
                
                // get device info values
//                 api.GetDeviceName(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceId(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceAddress(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceLatLong(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceIndoorLocation(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceBattery(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceType(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceUsername(interactionLifecycleEvent.SourceUri, false)
//                 api.GetDeviceLocationEnabled(interactionLifecycleEvent.SourceUri, false)
                
                // set device name
//                 api.SetDeviceName(interactionLifecycleEvent.SourceUri, "optimus prime")
//                 deviceName := api.GetDeviceName(interactionLifecycleEvent.SourceUri, false)
//                 api.Say(interactionLifecycleEvent.SourceUri, deviceName, "")
                
                // enable location
//                 api.EnableLocation(interactionLifecycleEvent.SourceUri)
//                 loc := api.GetDeviceLocationEnabled(interactionLifecycleEvent.SourceUri, false)
//                 fmt.Println("set location to true, got", loc)
//                 
//                 api.DisableLocation(interactionLifecycleEvent.SourceUri)
//                 loc2 := api.GetDeviceLocationEnabled(interactionLifecycleEvent.SourceUri, false)
//                 fmt.Println("set location to false, got", loc2)
                
                // user profile
//                 api.SetUserProfile(interactionLifecycleEvent.SourceUri, "test profile", false)
//                 profile := api.GetUserProfile(interactionLifecycleEvent.SourceUri)
//                 fmt.Println("user profile is ", profile)
                
                // set channel
//                 api.SetChannel(interactionLifecycleEvent.SourceUri, "User 40", false, false)
                
                // set device mode
//                 api.SetDeviceMode(interactionLifecycleEvent.SourceUri, sdk.DEVICE_MODE_NONE)
                
                // timers
//                 api.SetTimer(sdk.TIMEOUT_TIMER_TYPE, "first timer", 5, sdk.SECS_TIMEOUT_TYPE)
//                 api.SetTimer(sdk.TIMEOUT_TIMER_TYPE, "second timer", 10, sdk.SECS_TIMEOUT_TYPE)

                // restart & power down
//                 api.RestartDevice(interactionLifecycleEvent.SourceUri)
//                 api.PowerDownDevice(interactionLifecycleEvent.SourceUri)
                
//                 api.Terminate()
            }
        })
        
        api.OnButton(func(buttonEvent sdk.ButtonEvent) {
            fmt.Println("button pressed", buttonEvent.Button, buttonEvent.Taps, buttonEvent.SourceUri)
            api.Terminate()
        })
        
        api.OnTimerFired(func(timerFiredEvent sdk.TimerFiredEvent) {
            fmt.Println("timer fired! name ", timerFiredEvent.Name)
        })
    })
    
    sdk.InitializeRelaySdk(port)
}
