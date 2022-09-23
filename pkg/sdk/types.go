// Copyright © 2022 Relay Inc.

package sdk

type Event string
const(
    ERROR Event = "error"
    START = "start"
    STOP = "stop"
    INTERACTION_LIFECYCLE = "interaction_lifecycle"
    BUTTON = "button"
    TIMER = "timer"
    SPEECH = "speech"
    NOTIFICATION = "notification"
    INCIDENT = "incident"
    PROMPT = "prompt"
    PROMPT_START = "prompt_start"
    PROMPT_STOP = "prompt_stop"

    // The device we called is ringing. We are waiting for them to answer.
    // This event can occur on the caller.
    CALL_RINGING = "call_ringing"

    // A call attempt that was ringing, progressing, or incoming is now fully
    // connected. This event can occur on both the caller and the callee.
    CALL_CONNECTED = "call_connected"

    // A call that was once connected has become disconnected. This event can
    // occur on both the caller and the callee.
    CALL_DISCONNECTED = "call_disconnected"

    // A call failed to get connected. This event can occur on both the caller
    // and the callee.
    CALL_FAILED = "call_failed"

    // The device is receiving an inbound call request. This event can occur
    // on the callee.
    CALL_RECEIVED = "call_received"

    // There is a request to make an outbound call. This event can occur on
    // the caller after using the "Call X" voice command on the Assistant.
    CALL_START_REQUEST = "call_start_request"

    TIMER_FIRED = "timer_fired"
)

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

type MessageType string 
const (
    EVENT = "event"
    RESPONSE = "response"
)

type Language string 
const (
    ENGLISH = `en-US`
    GERMAN = `de-DE`
    SPANISH = `es-ES`
    FRENCH = `fr-FR`
    ITALIAN = `it-IT`
    RUSSIAN = `ru-RU`
    SWEDISH = `sv-SE`
    TURKISH = `tr-TR`
    HINDI = `hi-IN`
    ICELANDIC = `is-IS`
    JAPANESE = `ja-JP`
    KOREAN = `ko-KR`
    POLISH = `pl-PK`
    PORTUGUESE = `pt-BR`
    NORWEGIAN = `nb-NO`
    DUTCH = `nl-NL`
    CHINESE = `zh`
    ARABIC = `ar`
    VIETNAMESE = `vi-VN`
    INDONESIAN = `id-ID`
    FILIPINO = `fil-PH`
    DANISH = `da-DK`
    CZECH = `cs-CZ`
    GUJURATI = `gu-IN`
    HUNGARIAN = `hu-HU`
    TAMIL = `ta-IN`
    UKRANIAN = `uk-UA`
    SLOVAK = `sk-SK`
    ROMANIAN = `ro-RO`
    PUNJABI = `pa-IN`
    MALAY = `ms-MY`
    BENGALI = `bn-IN`
    GREEK = `el-GR`
    KANNADA = `kn-IN`
    FINNISH = `fi-FI`
)

type TimerType string
const (
    TIMEOUT_TIMER_TYPE = "timeout"
    INTERVAL_TIMER_TYPE = "interval"
)

type TimeoutType string
const (
    MS_TIMEOUT_TYPE = "ms"
    SECS_TIMEOUT_TYPE = "secs"
    MINS_TIMEOUT_TYPE = "mins"
    HRS_TIMEOUT_TYPE = "hrs"
)

type DeviceInfoQuery string
const (
    DEVICE_INFO_QUERY_NAME = "name"
    DEVICE_INFO_QUERY_ID = "id"
    DEVICE_INFO_QUERY_ADDRESS = "address"
    DEVICE_INFO_QUERY_LATLONG = "latlong"
    DEVICE_INFO_QUERY_INDOOR_LOCATION = "indoor_location"
    DEVICE_INFO_QUERY_BATTERY = "battery"
    DEVICE_INFO_QUERY_TYPE = "type"
    DEVICE_INFO_QUERY_USERNAME = "username"
    DEVICE_INFO_QUERY_LOCATION_ENABLED = "location_enabled"
)

// EVENTS 

// event structs are exported, but their _type and _id are not

type StartEvent struct {
    _type string `json:"_type"`
    Trigger Trigger 
}

type Trigger struct {
    Type TriggerType
    Args TriggerArgs
}

type NotificationOptions struct {
    priority NotificationPriority
    title string
    body string
    sound NotificationSound
}


type NotificationPriority string
const (
    NORMAL = `normal`
    HIGH = `high`
    CRITICAL = `critical`
)

type NotificationSound string
const (
    DEFAULT = `default`
    SOS = `sos`
)

type TriggerArgs struct {
    Phrase string `json:"phrase"`
    SourceUri string `json:"source_uri"`
}

type InteractionLifecycleEvent struct {
    _type string `json:"_type"`
    SourceUri string `json:"source_uri"`
    LifecycleType string `json:"type"`          // started, resumed
}

type PromptEvent struct {
    _type string `json:"_type"`
    SourceUri string `json:"source_uri"`
    PromptType string `json:"type"`             // started, stopped, resumed
}

type TimerFiredEvent struct {
    _type string `json:"_type"`
    Name string `json:"name"`
}

type TimerEvent struct {
    _type string `json:"_type"`
}

type ButtonEvent struct {
    _type string `json:"_type"`
    SourceUri string `json:"source_uri"`
    Button string `json:"button"`           // "action", "channel"
    Taps string `json:"taps"`               // "single", "double", "triple", "long"
}

type StopEvent struct {
    _type string `json:"_type"`
    Reason string `json:"reason"`
}

type SpeechEvent struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    SourceUri string `json:"source_uri"`
    ReuqestId string `json:"request_id"`
    Text string `json:"text"`
    Audio string `json:"audio"`
    Lang string `json:"lang"` 
}


// REQUEST/RESPONSE

// requests are not exported, because they are created internally, and their _type and _id fields are exported so they can be json encoded
// responses are exported, but their _type and _id are not, because they are not needed by user

type startInteractionRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Targets map[string][]string `json:"_target"`
    Name string `json:"name"`
    //Options ??
}

type StartInteractionResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
    SourceUri string `json:"source_uri"`
}

type endInteractionRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Targets map[string][]string `json:"_target"`
    //Options ??
}

type EndInteractionResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
    SourceUri string `json:"source_uri"`
}

type sayRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Text string `json:"text"`
    Lang Language `json:"lang"`
}

type SayResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`   
    CorrelationId string `json:"id"`
}

type listenRequest struct {
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Type string `json:"_type"`
    ReqestId string `json:"request_id"`
    Phrases []string `json:"phrases"`
    Transcribe bool `json:"transcribe"`
    Timeout int `json:"timeout"`
    AltLang string `json:"alt_lang"`
}

type ListenResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type translateRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Text string `json:"text"`
    FromLang Language `json:"from_lang"`
    ToLang Language `json:"to_lang"`
}

type TranslateResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    Text string `json:"text"`
}

type logAnalyticsEventRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Content string `json:"content"`
    ContentType string `json:"content_type"`
    Category string `json:"category"`
    DeviceUri string `json":"device_uri,omitempty"`
}

type LogAnalyticsEventResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type setVarRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Name string `json:"name"`
    Value string `json:"value"`
}

type SetVarResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    Name string `json:"name"`
    IType string `json:"type"`
    Value string `json:"value"`
}

type unsetVarRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Name string `json:"name"`
}

type UnsetVarResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type getVarRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Name string `json:"name"`
}

type GetVarResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    Value string `json:"value"`
}

type playRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Filename string `json:"filename"`
}

type PlayResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
    CorrelationId string `json:"id"`
}

type stopPlaybackRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Ids []string `json:"ids"`
}

type StopPlaybackResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type inboxCountRequest struct {
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Type string `json:"_type"`
}

type InboxCountResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    Count string `json:"count"`
}

type playInboxMessagesRequest struct {
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Type string `json:"_type"`
}

type PlayInboxMessagesResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type setHomeChannelStateRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Enabled bool `json:"enabled"`
}

type SetHomeChannelStateResponse struct {
    _type string `json:"_type"`
    Id string `json:"_id"`

}

type setTimerRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    TimerType TimerType `json:"type"`
    Name string `json:"name"`
    Timeout uint64 `json:"timeout"`
    TimeoutType TimeoutType `json:"timeout_type"`
}

type SetTimerResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type clearTimerRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Name string `json:"name"`
}

type ClearTimerResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type startTimerRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Timeout int `json:"timeout"`
}

type StartTimerResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type stopTimerRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
}

type StopTimerResponse struct {
    _id string `json"_id"`
    _type string `json:"_type"`
}

type createIncidentRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    IncidentType string `json:"type"`
    OriginatorUri string `json:"originator_uri"`
}

type CreateIncidentResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    IncidentId string `json:"incident_id"`
}

type resolveIncidentRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    IncidentId string `json:"incident_id"`
    Reason string `json:"reason"`
}

type ResolveIncidentResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type groupQueryRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    GroupUri string `json:"group_uri"`
    Query string `json:"query"`
}

type GroupQueryResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    MemberUris []string `json:"member_uris"`
    IsMember bool `json:"is_member"`
}

type LedEffect string 
const (
    LED_RAINBOW = "rainbow"
    LED_ROTATE = "rotate"
    LED_FLASH = "flash"
    LED_BREATHE = "breathe"
    LED_STATIC = "static"
    LED_OFF = "off"
)

type setLedRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Effect LedEffect `json:"effect"`
    Args LedInfo `json:"args"`
        
}

type LedInfo struct {
    Rotations int64 `json:"rotations,omitempty"`
    Count int64 `json:"count,omitempty"`
    Duration uint64 `json:"duration,omitempty"`
    RepeatDelay uint64 `json:"repeat_delay,omitempty"`
    PatternRepeats uint64 `json:"pattern_repeats,omitempty"`
    Colors LedColors `json:"colors,omitempty"`
}

type LedColors struct {
    Ring string `json:"ring,omitempty"`
    Led1 string `json:"1,omitempty"`
    Led2 string `json:"2,omitempty"`
    Led3 string `json:"3,omitempty"`
    Led4 string `json:"4,omitempty"`
    Led5 string `json:"5,omitempty"`
    Led6 string `json:"6,omitempty"`
    Led7 string `json:"7,omitempty"`
    Led8 string `json:"8,omitempty"`
    Led9 string `json:"9,omitempty"`
    Led10 string `json:"10,omitempty"`
    Led11 string `json:"11,omitempty"`
    Led12 string `json:"12,omitempty"`
    Led13 string `json:"13,omitempty"`
    Led14 string `json:"14,omitempty"`
    Led15 string `json:"15,omitempty"`
    Led16 string `json:"16,omitempty"`
}

func setLedColors(index string, color string) LedColors {
    if(index == "1") {
        return LedColors{Led1: color}
    }
    if(index == "2") {
        return LedColors{Led2: color}
    }
    if(index == "3") {
        return LedColors{Led3: color}
    }
    if(index == "4") {
        return LedColors{Led4: color}
    }
    if(index == "5") {
        return LedColors{Led5: color}
    }
    if(index == "6") {
        return LedColors{Led6: color}
    }
    if(index == "7") {
        return LedColors{Led7: color}
    }
    if(index == "8") {
        return LedColors{Led8: color}
    }
    if(index == "9") {
        return LedColors{Led9: color}
    }
    if(index == "10") {
        return LedColors{Led10: color}
    }
    if(index == "11") {
        return LedColors{Led11: color}
    }
    if(index == "12") {
        return LedColors{Led12: color}
    }
    if(index == "13") {
        return LedColors{Led13: color}
    }
    if(index == "14") {
        return LedColors{Led14: color}
    }
    if(index == "15") {
        return LedColors{Led15: color}
    }
    if(index == "16") {
        return LedColors{Led16: color}
    }
    return LedColors{Led1: color}
}

type SetLedResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type vibrateRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Pattern []uint64 `json:"pattern"`
}

type VibrateResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type sendNotificationRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Originator string `json:"originator"`
    IType string `json:"type"`
    Name string `json:"name"`
    Text string `json:"text"`
    ITarget map[string][]string `json:"target"`
    PushOptions NotificationOptions `json:"push_opts"`

}

type SendNotificationResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type getDeviceInfoRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Query DeviceInfoQuery `json:"query"`
    Refresh bool `json:"refresh"`
}

type GetDeviceInfoResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
    Name string `json:"name"`
    Id string `json:"id"`
    Address string `json:"address"`
    LatLong []float64 `json:"latlong"`
    IndoorLocation string `json:"indoor_location"`
    Battery uint64 `json:"battery"`
    Type string `json:"type"`
    Username string `json:"username"`
    LocationEnabled bool `json:"location_enabled"`
}

type SetDeviceInfoType string
const (
    SET_DEVICE_INFO_LABEL = "label"
    SET_DEVICE_INFO_LOCATION_ENABLED = "location_enabled"
)

type setDeviceInfoRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Field SetDeviceInfoType `json:"field"`
    Value string `json:"value"`
}

type SetDeviceInfoResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type setUserProfileRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Username string `json:"username"`
    Force bool `json:"force"`
}

type SetUserProfileResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type setChannelRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    ChannelName string `json:"channel_name"`
    SuppressTTS bool `json:"suppress_tts"`
    DisableHomeChannel bool `json:"disable_home_channel"`
}

type SetChannelResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type setDeviceModeRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Mode DeviceMode `json:"mode"`
}

type SetDeviceModeResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type DeviceMode string
const (
    DEVICE_MODE_PANIC = "panic"
    DEVICE_MODE_ALARM = "alarm"
    DEVICE_MODE_NONE = "none"
)

type devicePowerOffRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Restart bool `json:"restart"`
}

type DevicePowerOffResponse struct {
    _type string `json:"_type"`
    _id string `json:"_id"`
}

type placeCallRequest struct {
    Id string `json:"_id"`
    Type string `json:"_type"`
    Target map[string][]string `json:"_target"`
    Uri string `json:"uri"`
}

type PlaceCallResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
    call_id string `json:"call_id"`
}

type hangupCallRequest struct {
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Type string `json:"_type"`
    CallId string `json:"call_id"`
}

type HangupCallResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type answerRequest struct {
    Id string `json:"_id"`
    Target map[string][]string `json:"_target"`
    Type string `json:"_type"`
    CallId string `json:"call_id"`
}

type AnswerResponse struct {
    _id string `json:"_id"`
    _type string `json:"_type"`
}

type terminateRequest struct {
    Type string `json:"_type"`
    Id string `json:"_id"`
}
