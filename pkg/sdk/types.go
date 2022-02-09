
package sdk

type Event string
const(
    ERROR Event = "error"
    START = "start"
    STOP = "stop"
    INTERACTION_LIFECYCLE = "interaction_lifecycle"
    BUTTON = "button"
    TIMER = "timer"
    NOTIFICATION = "notification"
    INCIDENT = "incident"
    PROMPT_START = "prompt_start"
    PROMPT_STOP = "prompt_stop"
    CALL_RINGING = "call_ringing"
    CALL_CONNECTED = "call_connected"
    CALL_DISCONNECTED = "call_disconnected"
    CALL_FAILED = "call_failed"
    CALL_RECEIVED = "call_received"
    CALL_START_REQUEST = "call_start_request"
)

//{"_type":"wf_api_start_event","trigger":{"args":{"phrase":[104,101,108,108,111],"source_uri":[117,114,110,58,114,101,108,97,121,45,114,101,115,111,117,114,99,101,58,110,97,109,101,58,100,101,118,105,99,101,58,97,108,112,104,97]},"type":"phrase"}}
type StartEvent struct {
    Type string `json:"_type"`
    Trigger Trigger 
}

type Trigger struct {
    Type TriggerType
    Args TriggerArgs
}

type TriggerType string
const (
    PHRASE_TRIGGER = "phrase"
    BUTTON_TRIGGER = "button"
    HTTP_TRIGGER = "http"
	OTHER_TRIGGER = "other"
	NFC_TRIGGER = "nfc"
	CALENDAR_TRIGGER = "calendar"
	GEOFENCE_TRIGGER = "geofence"
	TELEPHONY_TRIGGER = "telephony"
)

type TriggerArgs struct {
    // TODO remove these when ibot serialization is fixed
    UnparsedPhrase []interface{} `json:"phrase"`
    UnparsedSourceUri []interface{} `json:"source_uri"`
    Phrase string
    SourceUri string
}

type StartInteractionRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Targets map[string][]string `json:"_target"`
    Name string `json:"name"`
    //Options ??
}

type StartInteractionResponse struct {
    Type string `json:"type"`
    SourceUri string `json:"source_uri"`
}

type InteractionLifecycleEvent struct {
    Type string `json:"type"`
    SourceUri string `json:"source_uri"`
}
// {"_type":"wf_api_interaction_lifecycle_event","source_uri":"urn:relay-resource:name:interaction:testing?device=urn%3Arelay-resource%3Aname%3Adevice%3Aalpha","type":"started"}
// {"_type":"wf_api_interaction_lifecycle_event","source_uri":"urn:relay-resource:name:interaction:testing?device=urn%3Arelay-resource%3Aname%3Adevice%3Aalpha","type":"resumed"}

type PromptEvent struct {
    Type string `json:"type"`
    SourceUri string `json:"source_uri"`
}
// {"_type":"wf_api_prompt_event","source_uri":"urn:relay-resource:name:interaction:testing?device=urn%3Arelay-resource%3Aname%3Adevice%3Aalpha","type":"stopped"}

type SayRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Text string `json:"text"`
    Lang string `json:"lang"`
}

type SayResponse struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    CorrelationId string `json:"id"`
}
// {"_id":"banana","_type":"wf_api_say_response","id":"banana"}

type TerminateRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
}

type StopEvent struct {
    Type string `json:"_type"`
    Reason string `json:"reason"`
}
// {"_type":"wf_api_stop_event","reason":"normal"}


 