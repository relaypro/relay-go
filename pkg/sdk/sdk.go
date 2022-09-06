// Copyright © 2022 Relay Inc.

package sdk

import (
    "net/http"
    "encoding/json"
    "regexp"
    log "github.com/sirupsen/logrus"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// this should return an interface that has a workflow() function that they can pass their workflow implementations to
func InitializeRelaySdk(port string) {
    log.Info("starting http server on", port)
    
    // use gorilla mux router
    r := mux.NewRouter()
    
    r.HandleFunc("/{workflowname}", handleWs)
    
    http.ListenAndServe(port, r)
}

func AddWorkflow(workflowName string, fn func(api RelayApi)) {
    // here we just register the wf by name, when a ws connects it will call the ws function passing the websocket in
    workflowMap[workflowName] = fn
    log.Info("Added workflow named", workflowName, "map is ", workflowMap)
}

func handleWs(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    wfName := vars["workflowname"]
    log.Debug("workflow name requested:", wfName)
    
    wfFunc, ok := workflowMap[wfName]
    if !ok {
        log.Debug("no workflow named ", wfName, "is registered", workflowMap)
        return
    }
        
    conn, upgradeErr := upgrader.Upgrade(w, r, nil)

    if upgradeErr != nil {
        log.Debug("upgrade error", upgradeErr)
        return
    }
    
    // at this point, a device has connected, and we have the wf name and the wf function that was registered
    // run the wf func by passing the ws connection to it
    // the name of the workflow is in the path that was requested
    
    // need to pass the ws conn, and the relayapi interface? do we really need anything except the conn?
    
    // start an async function to run the wf and handle the ws 
    wfInst := &workflowInstance{
        WebsocketConnection: conn, 
        WorkflowFn: wfFunc, 
        Pending: make(map[string]*Call), 
        EventChannel: make(chan EventWrapper, 100),
    }
    go startWorkflow(wfInst, wfName)
    
}

func startWorkflow(wfInst *workflowInstance, workflowName string) {
    // this thread blocks in 2 places, when waiting for a message to come over the ws, or when waiting for a response to 
    // a request that was sent. ws listening in done on a separate coroutine, event messages are sent to this coroutine,
    // and response messages are handled on the listening corouting to complete the call object since this coroutine will
    // be blocked waiting for the response
    
    // this looks weird, but the wfInst struct holds the user's workflow function, and we pass the wfInst to it because it implements the RelayApi interface that the workflowFn accepts
    // call the user defined wf function, passing the RelayApi interface to it (which is implemented on the workflowInstance struct)
    wfInst.WorkflowFn(wfInst)
    
    // listen for ws messages in a coroutine so we can receive responses while blocking on this coroutine
    go wfInst.receiveWs()

    log.Info("Workflow instance started for ", workflowName)

    // loop forever handling events and responses    
    var err error 
    for err == nil {
        select {
            case eventWrapper := <-wfInst.EventChannel:
                err = wfInst.handleEvent(eventWrapper)
        }
    }
    log.Debug("exiting, err is ", err)
    log.Info("Workflow instance terminating, reason: ", err)
}

var eventRegex = regexp.MustCompile(`^wf_api_(.+)_event$`)
var responseRegex = regexp.MustCompile(`^wf_api_(.+)_response$`)

func parseMessage(msg []byte) (map[string]interface{}, Event, string) {
    var parsedMsg map[string]interface{}
    json.Unmarshal(msg, &parsedMsg)
    
    // match _type against regexes to find out if it's an event or a response
    var matches []string
    var messageType string          // "event" or "response"
    msgType := parsedMsg["_type"].(string)
    if eventRegex.MatchString(msgType) {
        matches = eventRegex.FindStringSubmatch(msgType)
        messageType = EVENT
    } else {
        matches = responseRegex.FindStringSubmatch(msgType)
        messageType = RESPONSE
    }
    
    return parsedMsg, Event(matches[1]), messageType
}

func convert(arr []interface{}) string {
    var bytes []byte = make([]byte, len(arr))
    for i, v := range arr {
        var f float64 = v.(float64)
        bytes[i] = byte(f)
    }
    return string(bytes)
}
