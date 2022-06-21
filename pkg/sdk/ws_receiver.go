// Copyright © 2022 Relay Inc.

package sdk

import (
    "fmt"
    "errors"
)

func (wfInst *workflowInstance) receiveWs() {
    defer wfInst.WebsocketConnection.Close()

    var err error 
    for err == nil {
        // Read message from websocket connection
        _, msg, err := wfInst.WebsocketConnection.ReadMessage()
        if err != nil {
            if wfInst.StopReason != "" {
                fmt.Println("websocket closed with reason:", wfInst.StopReason)
                return
            } else {
                fmt.Println("Error reading message from websocket:", err, msg)
                return
            }
        }
        
        // messages are either events or responses to requests we sent
        parsedMsg, eventName, messageType := parseMessage(msg)
        eventWrapper := EventWrapper{ParsedMsg: parsedMsg, Msg: msg, EventName: eventName}
        if messageType == "event" {
            // send events to event channel
            select {
                case wfInst.EventChannel <- eventWrapper:
                default:
                    fmt.Println("Error, can't send to event channel")
                    return
            }
        } else if messageType == "response" {
            // pair with callback
            err = wfInst.handleResponse(EventWrapper{ParsedMsg: parsedMsg, Msg: msg, EventName: eventName})
            if err != nil {
                fmt.Println("Error from response handler ", err)
                return
            }
        }
    }
    fmt.Println("error received from websocket", err, "quitting")
}

func (wfInst *workflowInstance) handleResponse(eventWrapper EventWrapper) error {
    fmt.Println("handling response for ", eventWrapper.ParsedMsg)
    
    // find the matching request and complete the call
    id := eventWrapper.ParsedMsg["_id"].(string)
    wfInst.Mutex.Lock()
    call := wfInst.Pending[id]
    delete(wfInst.Pending, id)
    wfInst.Mutex.Unlock()
    
    if call == nil {
        fmt.Println("no call associated with response", eventWrapper.ParsedMsg)
        return errors.New("no pending request found")
    }
    
    call.EventWrapper = eventWrapper
    call.Res = eventWrapper.ParsedMsg
    call.Done <- true
    
    return nil
}
