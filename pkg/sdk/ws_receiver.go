// Copyright © 2022 Relay Inc.

package sdk

import (
    log "github.com/sirupsen/logrus"
)

func (wfInst *workflowInstance) receiveWs() {
    defer wfInst.WebsocketConnection.Close()

    var err error 
    for err == nil {
        // Read message from websocket connection
        _, msg, err := wfInst.WebsocketConnection.ReadMessage()
        if err != nil {
            if wfInst.StopReason != "" {
                log.Debug("websocket closed with reason:", wfInst.StopReason)
                return
            } else {
                log.Debug("Error reading message from websocket:", err, msg)
                return
            }
        }
        
        // messages are either events or responses to requests we sent
        parsedMsg, eventName, messageType := parseMessage(msg)
        eventWrapper := EventWrapper{ParsedMsg: parsedMsg, Msg: msg, EventName: eventName}
        
        // including speech event so that it can be passed to handleResponse for a listen API call
        if messageType == "response" || eventName == "speech" {
            // pair with callback
            err = wfInst.handleResponse(EventWrapper{ParsedMsg: parsedMsg, Msg: msg, EventName: eventName})
            if err != nil {
                log.Debug("Error from response handler ", err)
                return
            }
        } else if messageType == "event" {
            // send events to event channel
            select {
                case wfInst.EventChannel <- eventWrapper:
                    if eventName == "prompt" && eventWrapper.ParsedMsg["type"].(string) == "stopped" {
                        streamingComplete = true
                    }

                default:
                    log.Debug("Error, can't send to event channel")
                    return
            }
        } 
    }
    log.Debug("error received from websocket", err, "quitting")
}

func (wfInst *workflowInstance) handleResponse(eventWrapper EventWrapper) error {
    log.Debug("handling response for ", eventWrapper.ParsedMsg)
    // find the matching request and complete the call. If the type is a speech event, it will contain a "request_id" instead of "_id".  This
    // request_id will correspond to the listen request id, if a listen was called.
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
        call.EventWrapper = eventWrapper
        call.Res = eventWrapper.ParsedMsg
        call.Done <- true
    }
    return nil
}
