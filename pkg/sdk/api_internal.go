// Copyright © 2022 Relay Inc.

package sdk

import (
    "fmt"
    "math/rand"
    "encoding/hex"
    "time"
    "errors"
    "encoding/json"
)

func (wfInst *workflowInstance) sendRequest(msg interface{}) {
    err := wfInst.WebsocketConnection.WriteJSON(&msg)
    if err != nil {
        fmt.Println("error sending message", err)
    }
}

func (wfInst *workflowInstance) sendAndReceiveRequest(msg interface{}, id string) *Call {
    // mutex is used to synchronize access to Pending map, and to lock the websocket write call
    wfInst.Mutex.Lock()
    call := &Call{Req: msg, Done: make(chan bool, 100)}
    wfInst.Pending[id] = call
    wfInst.Mutex.Unlock()
    
    err := wfInst.WebsocketConnection.WriteJSON(&msg)
    if err != nil {
        fmt.Println("error sending message", err)
        // remove the pending call
        wfInst.Mutex.Lock()
        delete(wfInst.Pending, id)
        wfInst.Mutex.Unlock()
    }
    fmt.Println("Sent request:", msg)
    
    // here we block to receive from the call's channel
    select {
        case <-call.Done:
        case <-time.After(10 * time.Second):
            fmt.Println("Request timed out")
            call.Error = errors.New("request timeout")
    }
    return call
}

func (wfInst *workflowInstance) handleEvent(eventWrapper EventWrapper) error {
    fmt.Println("Handling event of type", eventWrapper.ParsedMsg["_type"])
    // call the appropriate handler function, if it was set by the user implementation
    switch eventWrapper.EventName {
        case "start":
            var params StartEvent
            json.Unmarshal(eventWrapper.Msg, &params)

            if wfInst.OnStartHandler != nil {
                wfInst.OnStartHandler(params)
            } else {
                fmt.Println("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "interaction_lifecycle":
            fmt.Println("interaction lifecycle event:", string(eventWrapper.Msg))
            var params InteractionLifecycleEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnInteractionLifecycleHandler != nil {
                wfInst.OnInteractionLifecycleHandler(params)
            } else {
                fmt.Println("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "prompt":
            var params PromptEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            fmt.Println("prompt event: ", params)
            if wfInst.OnPromptHandler != nil {
                wfInst.OnPromptHandler(params)
            } else {
                fmt.Println("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "button":
            fmt.Println("button event", string(eventWrapper.Msg))
            var params ButtonEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnButtonHandler != nil {
                wfInst.OnButtonHandler(params)
            } else {
                fmt.Println("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "stop":
            fmt.Println("received stop event", string(eventWrapper.Msg))
            var params StopEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            wfInst.StopReason = params.Reason
        case "timer_fired":
            fmt.Println("received timer fired event")
            var params TimerFiredEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnTimerFiredHandler != nil {
                wfInst.OnTimerFiredHandler(params)
            } else {
                fmt.Println("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        default:
            fmt.Println("UNKNOWN EVENT ", eventWrapper.ParsedMsg);
    }
    return nil
}

func makeId() string {
    r := make([]byte, 16)
    rand.Read(r)
    return hex.EncodeToString(r)
}

func makeTargetMap(sourceUri string) map[string][]string {
     return map[string][]string {
         "uris": []string{sourceUri},
     }   
}
