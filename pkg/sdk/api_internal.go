// Copyright © 2022 Relay Inc.

package sdk

import (
    "fmt"
    "math/rand"
    log "github.com/sirupsen/logrus"
    "encoding/hex"
    "time"
    "errors"
    "encoding/json"
)

// boolean variable used to keep track of whether or not streaming is complete on the device.  Mainly used for the functions
// SayAndWait and PlayAndWait, which require streaming to complete on the device before continuing through the workflow.
var streamingComplete bool

func (wfInst *workflowInstance) sendRequest(msg interface{}) {
    err := wfInst.WebsocketConnection.WriteJSON(&msg)
    if err != nil {
        log.Error("error sending message ", err)
    }
}

func (wfInst *workflowInstance) sendAndReceiveRequest(msg interface{}, id string) *Call {
    // does not require streaming to complete on the device before continuing
    streamingComplete = true
    // mutex is used to synchronize access to Pending map, and to lock the websocket write call
    wfInst.Mutex.Lock()
    call := &Call{Req: msg, Done: make(chan bool, 100)}
    wfInst.Pending[id] = call
    wfInst.Mutex.Unlock()
    
    err := wfInst.WebsocketConnection.WriteJSON(&msg)
    if err != nil {
        log.Error("error sending message ", err)
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
            log.Debug("Request timed out")
            call.Error = errors.New("request timeout")
    }
    return call
}


func (wfInst *workflowInstance) sendAndReceiveRequestWait(msg interface{}, id string) *Call {
    // mutex is used to synchronize access to Pending map, and to lock the websocket write call
    wfInst.Mutex.Lock()
    call := &Call{Req: msg, Done: make(chan bool, 100)}
    wfInst.Pending[id] = call
    wfInst.Mutex.Unlock()
    
    err := wfInst.WebsocketConnection.WriteJSON(&msg)
    if err != nil {
        log.Error("error sending message ", err)
        // remove the pending call
        wfInst.Mutex.Lock()
        delete(wfInst.Pending, id)
        wfInst.Mutex.Unlock()
    }
    fmt.Println("Sent request:", msg)
    // here we block to receive from the call's channel
    select {
        // once the call is done, wait until your receive a prompt event before returning the call
        case <-call.Done:
            // you need to wait for streaming to complete on the device before the next function call
            streamingComplete = false
            startTime := time.Now()
            log.Debug("Waiting for prompt stopped")
            for !streamingComplete {
                if(time.Since(startTime).Seconds() >= 30) {
                    fmt.Println("Timed out waiting for prompt event")
                    break
                }
            }
        case <-time.After(10 * time.Second):
            log.Debug("Request timed out")
            call.Error = errors.New("request timeout")
    }
    return call
}

func (wfInst *workflowInstance) handleEvent(eventWrapper EventWrapper) error {
    log.Debug("Handling event of type ", eventWrapper.ParsedMsg["_type"])
    // call the appropriate handler function, if it was set by the user implementation
    switch eventWrapper.EventName {
        case "start":
            var params StartEvent
            json.Unmarshal(eventWrapper.Msg, &params)

            if wfInst.OnStartHandler != nil {
                wfInst.OnStartHandler(params)
            } else {
                log.Debug("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "interaction_lifecycle":
            log.Debug("interaction lifecycle event:", string(eventWrapper.Msg))
            var params InteractionLifecycleEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnInteractionLifecycleHandler != nil {
                wfInst.OnInteractionLifecycleHandler(params)
            } else {
                log.Debug("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "prompt":
            var params PromptEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            log.Debug("prompt event: ", params)
            if wfInst.OnPromptHandler != nil {
                wfInst.OnPromptHandler(params)
            } else {
                log.Debug("ignoring event", eventWrapper.EventName, "no handler registered")                
            }
        case "button":
            log.Debug("button event ", string(eventWrapper.Msg))
            var params ButtonEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnButtonHandler != nil {
                wfInst.OnButtonHandler(params)
            } else {
                log.Debug("ignoring event ", eventWrapper.EventName, "no handler registered")                
            }
        case "stop":
            log.Debug("received stop event ", string(eventWrapper.Msg))
            var params StopEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            wfInst.StopReason = params.Reason
        case "timer_fired":
            log.Debug("received timer fired event")
            var params TimerFiredEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnTimerFiredHandler != nil {
                wfInst.OnTimerFiredHandler(params)
            } else {
                log.Debug("ignoring event ", eventWrapper.EventName, " no handler registered")                
            }
        case "timer":
            log.Debug("received timer event ", string(eventWrapper.Msg))
            var params TimerEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if wfInst.OnTimerHandler != nil {
                wfInst.OnTimerHandler(params)
            } else {
                log.Debug("ignoring event ", eventWrapper.EventName, " no handler registered")
            }
        case "speech":
            log.Debug("received speech event ", string(eventWrapper.Msg))
            var params SpeechEvent
            json.Unmarshal(eventWrapper.Msg, &params)
            if(wfInst.OnSpeechHandler != nil) {
                wfInst.OnSpeechHandler(params)
            } else {
                log.Debug("ignoring event ", eventWrapper.EventName, " no handler registered")
            }
        default:
            log.Debug("UNKNOWN EVENT ", eventWrapper.ParsedMsg);
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
