package sdk

import (
    "fmt"
    "encoding/json"
    "github.com/gorilla/websocket"
)

var workflowMap map[string]func(api RelayApi) = make(map[string]func(api RelayApi))

type RelayApi interface {            // this is interface of your custom workflow, you implement this, then we call it and pass in the ws
    // assigning callbacks
    OnStart(fn func(startEvent StartEvent))
    OnInteractionLifecycle(func(interactionLifecycleEvent InteractionLifecycleEvent))
    OnPrompt(func(promptEvent PromptEvent))
    // api
    StartInteraction(sourceUri string)
    Say(sourceUri string, text string)
    Terminate()
/*  SetTimer()
    ClearTimer()
    RestartDevice()
    PowerDownDevice()
    Say()
*/
    // etc...
}

// This struct implements RelayApi below
type workflowInstance struct {
    WebsocketConnection *websocket.Conn
    WorkflowFn func(api RelayApi)
    
    // stores callback functions for each event type
    OnStartHandler func(startEvent StartEvent)
    OnInteractionLifecycleHandler func(interactionLifecycleEvent InteractionLifecycleEvent)
    OnPromptHandler func(promptEvent PromptEvent)
}


// interface function implementions
func (wf *workflowInstance) OnStart(fn func(startEvent StartEvent)) {
    fmt.Println("i'm the on function, setting callback for start event")
    
    // store the func that was passed in as a callback in a slice, then whenever the websocket sends us a matching event, call the callback
    wf.OnStartHandler = fn           // set the callback for this event type
}

func (wf *workflowInstance) OnInteractionLifecycle(fn func(interactionLifecycleEvent InteractionLifecycleEvent)) {
    fmt.Println("setting start lifecycle handler")
    wf.OnInteractionLifecycleHandler = fn
}

func (wf *workflowInstance) OnPrompt(fn func(promptEvent PromptEvent)) {
    fmt.Println("setting prompt event handler")
    wf.OnPromptHandler = fn
}

func (wf *workflowInstance) StartInteraction(sourceUri string) {
    fmt.Println("i guess i'll start an interaction now for", sourceUri)
    targets := map[string][]string {
        "uris": []string{sourceUri},
    }
    req := StartInteractionRequest{Type: "wf_api_start_interaction_request", Id: "banana", Targets: targets, Name: "testing"}
    str, _ := json.Marshal(req)
    sendMessage(wf.WebsocketConnection, str)
}

func (wf *workflowInstance) Say(sourceUri string, text string) {
    fmt.Println("saying ", text, "to", sourceUri)
    targets := map[string][]string {
        "uris": []string{sourceUri},
    }
    req := SayRequest{Type: "wf_api_say_request", Id: "banana", Target: targets, Text: text, Lang: "en-US"}
    str, _ := json.Marshal(req)
    sendMessage(wf.WebsocketConnection, str)
}

func (wf *workflowInstance) Terminate() {
    fmt.Println("terminating")
    req := TerminateRequest{Type: "wf_api_terminate_request", Id: "banana"}
    str, _ := json.Marshal(req)
    sendMessage(wf.WebsocketConnection, str)
}

func sendMessage(conn *websocket.Conn, msg []byte) {
    err := conn.WriteMessage(websocket.TextMessage, msg)
    if err != nil {
        fmt.Println("error sending message", err)
        return
    }
    fmt.Println("sent message ", msg)
}
