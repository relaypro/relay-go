// Copyright © 2022 Relay Inc.

package sdk

import (
    "fmt"
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
        fmt.Println("EVENT'S NAME", eventName)
        if eventName == "prompt" {
            promptReceived = true
        }
        if messageType == "event"  && eventName != "speech"{
            // send events to event channel
            select {
                case wfInst.EventChannel <- eventWrapper:
                default:
                    fmt.Println("Error, can't send to event channel")
                    return
            }
        } else if messageType == "response" || eventName == "speech" {
            fmt.Println("PROMPT IN RECEIVEWS", promptReceived)
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
    fmt.Println("PROMPT: ", promptReceived)
    // find the matching request and complete the call 
    var id string
    if (eventWrapper.ParsedMsg["_type"].(string) == "wf_api_speech_event") {
        id = eventWrapper.ParsedMsg["request_id"].(string)
    } else {
        id = eventWrapper.ParsedMsg["_id"].(string)
    }
    if (eventWrapper.ParsedMsg["_type"].(string) != "wf_api_listen_response") {
        wfInst.Mutex.Lock()
        call := wfInst.Pending[id]
        delete(wfInst.Pending, id)
        wfInst.Mutex.Unlock()
        fmt.Println("Call: ", call)
        call.EventWrapper = eventWrapper
        fmt.Println("Event wrapper", eventWrapper.EventName)
        call.Res = eventWrapper.ParsedMsg
        call.Done <- true
    }
    return nil
}
