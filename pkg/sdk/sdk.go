package sdk

import (
    "fmt"
    "net/http"
    "encoding/json"
    "regexp"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

var port = ":5000"

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

// this should return an interface that has a workflow() function that they can pass their workflow implementations to
func InitializeRelaySdk() {
    // TODO do this in a thread so it doesn't block further calls?
    fmt.Println("starting http server on", port)
    
    // use gorilla mux router
    r := mux.NewRouter()
    
    r.HandleFunc("/ws/{workflowname}", handleWs)
//     r.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
//         http.ServeFile(w, r, "websockets.html")
//     })
    r.HandleFunc("/", handle)
    
    http.ListenAndServe(port, r)
}

func AddWorkflow(workflowName string, fn func(api RelayApi)) {
    // here we just register the wf by name, when a ws connects it will call the ws function passing the websocket in
    workflowMap[workflowName] = fn
    fmt.Println("Added workflow named", workflowName, "map is ", workflowMap)
}

func handleWs(w http.ResponseWriter, r *http.Request) {

    fmt.Println("received request", *r)

    vars := mux.Vars(r)
    wfName := vars["workflowname"]
    fmt.Println("workflow name requested:", wfName)
    
    wfFunc, ok := workflowMap[wfName]
    if !ok {
        fmt.Println("no workflow named ", wfName, "is registered", workflowMap)
        return
    }
    fmt.Println("Found workflow named ", wfName, wfFunc)
    
    // wf is a func(api RelayApi){}
    
    conn, upgradeErr := upgrader.Upgrade(w, r, nil)

    if upgradeErr != nil {
        fmt.Println("upgrade error", upgradeErr)
        return
    }
    
    // at this point, a device has connected, and we have the wf name and the wf function that was registered
    // run the wf func by passing the ws connection to it
    // the name of the workflow is in the path that was requested
    
    // need to pass the ws conn, and the relayapi interface? do we really need anything except the conn?
    
    // start an async function to run the wf and handle the ws 
    wfInst := workflowInstance{WebsocketConnection: conn, WorkflowFn: wfFunc}
    go startWorkflow(wfInst)
    
    
    
    

}

func handle(w http.ResponseWriter, r *http.Request) {
    fmt.Println("received request", *r)
    fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
}

func startWorkflow(wfInst workflowInstance) {
    
    // this looks weird, but the wfInst struct holds the user's workflow function, and we pass the wfInst to it because it implements the RelayApi interface that the workflowFn accepts
    // call the user defined wf function, passing the RelayApi interface to it (which is implemented on the workflowInstance struct)
    wfInst.WorkflowFn(&wfInst)
    fmt.Println("wfInst after calling user implementation", wfInst)
    
    fmt.Println("wf instance: ", wfInst)
    
    // test calling the on button callback
    //wfInst.OnHandlers["BUTTON"]()
    
    // do this in the workflow function so that we can pass these messages to the wf implementation
    for {
        // Read message from browser
        msgType, msg, err := wfInst.WebsocketConnection.ReadMessage()
        if err != nil {
            fmt.Println("Error reading message from websocket:", err, msgType, msg)
            return
        }

        // Print the message to the console
        fmt.Printf("received message from %s: %s\n", wfInst.WebsocketConnection.RemoteAddr(), string(msg))

        parsedMsg, event := parseMessage(msg)
                
        fmt.Println("got a message with event:", event, "message:", parsedMsg)
        // call the appropriate handler function, if it was set by the user implementation
        switch event {
            case "start":
                var params StartEvent
                json.Unmarshal(msg, &params)
                
                // TODO This won't be necessary once ibot fixes serialization
                // source uri and phrase are serialized as byte arrays not binary
                //fmt.Println("params for start event:", params)
                //fmt.Println("converted source uri", params.Trigger.Args.UnparsedSourceUri, "to", convert(params.Trigger.Args.UnparsedSourceUri))
                //fmt.Printf("type is %T", params.Trigger.Args.UnparsedSourceUri)
                params.Trigger.Args.SourceUri = convert(params.Trigger.Args.UnparsedSourceUri)
                params.Trigger.Args.Phrase = convert(params.Trigger.Args.UnparsedPhrase)
                
                if wfInst.OnStartHandler != nil {
                    wfInst.OnStartHandler(params)
                } else {
                    fmt.Println("ignoring event", event, "no handler registered")                
                }
            case "start_interaction":
                fmt.Println("started interaction", parsedMsg)
                // TODO what to do with these responses? need to pair them up with their requests to block until the resp is received?

            case "interaction_lifecycle":
                fmt.Println("interaction lifecycle event:", parsedMsg)
                var params InteractionLifecycleEvent
                json.Unmarshal(msg, &params)
                if wfInst.OnInteractionLifecycleHandler != nil {
                    wfInst.OnInteractionLifecycleHandler(params)
                } else {
                    fmt.Println("ignoring event", event, "no handler registered")                
                }
            case "prompt":
                var params PromptEvent
                json.Unmarshal(msg, &params)
                if wfInst.OnPromptHandler != nil {
                    wfInst.OnPromptHandler(params)
                } else {
                    fmt.Println("ignoring event", event, "no handler registered")                
                }
            case "say":
                fmt.Println("received say response", parsedMsg)
            case "stop":
                fmt.Println("received stop event", parsedMsg)
            default:
                fmt.Println("UNKNOWN EVENT ", event);
        }
    }
}



var eventRegex = regexp.MustCompile(`^wf_api_(.+)_event$`)
var responseRegex = regexp.MustCompile(`^wf_api_(.+)_response$`)

func parseMessage(msg []byte) (map[string]interface{}, Event) {
    var result map[string]interface{}
    json.Unmarshal(msg, &result)
    fmt.Println("parsed msg", result)
    
    var matches []string
    if eventRegex.MatchString(result["_type"].(string)) {
        matches = eventRegex.FindStringSubmatch(result["_type"].(string))
    } else {
        matches = responseRegex.FindStringSubmatch(result["_type"].(string))
    }
    fmt.Println("matches found", matches)
    
    return result, Event(matches[1])
}

func convert(arr []interface{}) string {
    var bytes []byte = make([]byte, len(arr))
    for i, v := range arr {
        var f float64 = v.(float64)
        bytes[i] = byte(f)
    }
    return string(bytes)
}
