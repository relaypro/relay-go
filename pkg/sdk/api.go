package sdk

import (
    "fmt"
    "sync"
    "encoding/json"
    "github.com/gorilla/websocket"
)

var workflowMap map[string]func(api RelayApi) = make(map[string]func(api RelayApi))

type RelayApi interface {            // this is interface of your custom workflow, you implement this, then we call it and pass in the ws
    // assigning callbacks
    OnStart(fn func(startEvent StartEvent))
    OnInteractionLifecycle(func(interactionLifecycleEvent InteractionLifecycleEvent))
    OnPrompt(func(promptEvent PromptEvent))
    OnTimerFired(func(timerFiredEvent TimerFiredEvent))
    OnButton(func(buttonEvent ButtonEvent))
    
    // api
    StartInteraction(sourceUri string) StartInteractionResponse
    SetTimer(timerType TimerType, name string, timeout uint64, timeoutType TimeoutType) SetTimerResponse
    ClearTimer(name string) ClearTimerResponse
    Say(sourceUri string, text string, lang string) SayResponse
    Play(sourceUri string, filename string) string
    StopPlayback(sourceUri string, ids []string) StopPlaybackResponse
    SwitchAllLedOn(sourceUri string, color string) SetLedResponse
    SwitchAllLedOff(sourceUri string) SetLedResponse
    Rainbow(sourceUri string, rotations int64) SetLedResponse
    Rotate(sourceUri string, color string) SetLedResponse
    Flash(sourceUri string, color string) SetLedResponse
    Breathe(sourceUri string, color string) SetLedResponse
    SetLeds(sourceUri string, effect LedEffect, args LedInfo) SetLedResponse
    Vibrate(sourceUri string, pattern []uint64) VibrateResponse
    GetDeviceName(sourceUri string, refresh bool) string
    GetDeviceId(sourceUri string, refresh bool) string
    GetDeviceAddress(sourceUri string, refresh bool) string
    GetDeviceLatLong(sourceUri string, refresh bool) []float64
    GetDeviceIndoorLocation(sourceUri string, refresh bool) string
    GetDeviceBattery(sourceUri string, refresh bool) uint64
    GetDeviceType(sourceUri string, refresh bool) string
    GetDeviceUsername(sourceUri string, refresh bool) string
    GetDeviceLocationEnabled(sourceUri string, refresh bool) bool
    SetDeviceName(sourceUri string, name string) SetDeviceInfoResponse
//     SetDeviceChannel(sourceUri string, channel string) SetDeviceInfoResponse
    EnableLocation(sourceUri string) SetDeviceInfoResponse
    DisableLocation(sourceUri string) SetDeviceInfoResponse
    SetUserProfile(sourceUri string, username string, force bool) SetUserProfileResponse
    SetChannel(sourceUri string, channelName string, suppressTTS bool, disableHomeChannel bool) SetChannelResponse

    SetDeviceMode(sourceUri string, mode DeviceMode) SetDeviceModeResponse
    
    RestartDevice(sourceUri string) DevicePowerOffResponse
    PowerDownDevice(sourceUri string) DevicePowerOffResponse
    Terminate()
    // -----




//     Translate()


// EnableHomeChannel(target: string|string[])
// DisableHomeChannel(target: string|string[])

    
    // -----
}

// This struct implements RelayApi below
type workflowInstance struct {
    WebsocketConnection *websocket.Conn
    Mutex sync.Mutex                  // no initialization, zero value is unlocked mutex. this must not be copied, always pass workflowInstance by pointer
    Pending map[string]*Call            // map of request ids to the call struct for response pairing
    WorkflowFn func(api RelayApi)
    
    EventChannel chan EventWrapper
    StopReason string
    
    // stores callback functions for each event type
    OnStartHandler func(startEvent StartEvent)
    OnInteractionLifecycleHandler func(interactionLifecycleEvent InteractionLifecycleEvent)
    OnPromptHandler func(promptEvent PromptEvent)
    OnButtonHandler func(buttonEvent ButtonEvent)
    OnTimerFiredHandler func(timerFiredEvent TimerFiredEvent)
}

// Call represents an active request
type Call struct {
	Req interface{}
	Res interface{}
	EventWrapper EventWrapper
	Done chan bool
	Error error
}

type EventWrapper struct {
    ParsedMsg map[string]interface{}
    Msg []byte
    EventName Event
}

// Callback Handlers

func (wfInst *workflowInstance) OnStart(fn func(startEvent StartEvent)) {
    // store the func that was passed in as a callback in a slice, then whenever the websocket sends us a matching event, call the callback
    wfInst.OnStartHandler = fn           // set the callback for this event type
}

func (wfInst *workflowInstance) OnInteractionLifecycle(fn func(interactionLifecycleEvent InteractionLifecycleEvent)) {
    wfInst.OnInteractionLifecycleHandler = fn
}

func (wfInst *workflowInstance) OnPrompt(fn func(promptEvent PromptEvent)) {
    wfInst.OnPromptHandler = fn
}

func (wfInst *workflowInstance) OnButton(fn func(buttonEvent ButtonEvent)) {
    wfInst.OnButtonHandler = fn
}

func (wfInst *workflowInstance) OnTimerFired(fn func(timerFiredEvent TimerFiredEvent)) {
    wfInst.OnTimerFiredHandler = fn
}


// API functions

func (wfInst *workflowInstance) StartInteraction(sourceUri string) StartInteractionResponse {
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := startInteractionRequest{Type: "wf_api_start_interaction_request", Id: id, Targets: target, Name: "testing"}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := StartInteractionResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) SetTimer(timerType TimerType, name string, timeout uint64, timeoutType TimeoutType) SetTimerResponse {
    id := makeId()
    req := setTimerRequest{Type: "wf_api_set_timer_request", Id: id, TimerType: timerType, Name: name, Timeout: timeout, TimeoutType: timeoutType}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SetTimerResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) ClearTimer(name string) ClearTimerResponse {
    id := makeId()
    req := clearTimerRequest{Type: "wf_api_clear_timer_request", Id: id, Name: name}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := ClearTimerResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) Say(sourceUri string, text string, lang string) SayResponse {
    if lang == "" {
        lang = "en-US"
    }
    fmt.Println("saying ", text, "to", sourceUri, "with lang", lang)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := sayRequest{Type: "wf_api_say_request", Id: id, Target: target, Text: text, Lang: lang}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SayResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) Play(sourceUri string, filename string) string {
    fmt.Println("playing file ", filename, "to", sourceUri)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := playRequest{Type: "wf_api_play_request", Id: id, Target: target, Filename: filename}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := PlayResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res.CorrelationId
}

func (wfInst *workflowInstance) StopPlayback(sourceUri string, ids []string) StopPlaybackResponse {
    fmt.Println("stopping playback for", ids)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := stopPlaybackRequest{Type: "wf_api_stop_playback_request", Id: id, Target: target, Ids: ids}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := StopPlaybackResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) SetLeds(sourceUri string, effect LedEffect, args LedInfo) SetLedResponse {
    fmt.Println("setting leds", effect, "with args", args)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := setLedRequest{Type: "wf_api_set_led_request", Id: id, Target: target, Effect: effect, Args: args}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SetLedResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) SwitchAllLedOn(sourceUri string, color string) SetLedResponse {
    return wfInst.SetLeds(sourceUri, LED_STATIC, LedInfo{Colors: LedColors{ Ring: color}})
}

func (wfInst *workflowInstance) SwitchAllLedOff(sourceUri string) SetLedResponse {
    return wfInst.SetLeds(sourceUri, LED_OFF, LedInfo{})
}

func (wfInst *workflowInstance) Rainbow(sourceUri string, rotations int64) SetLedResponse {
    return wfInst.SetLeds(sourceUri, LED_RAINBOW, LedInfo{Rotations: rotations})
}

func (wfInst *workflowInstance) Rotate(sourceUri string, color string) SetLedResponse {
    return wfInst.SetLeds(sourceUri, LED_ROTATE, LedInfo{Rotations: -1, Colors: LedColors{ Led1: color }})
}

func (wfInst *workflowInstance) Flash(sourceUri string, color string) SetLedResponse {
    return wfInst.SetLeds(sourceUri, LED_FLASH, LedInfo{Count: -1, Colors: LedColors{ Ring: color }})
}

func (wfInst *workflowInstance) Breathe(sourceUri string, color string) SetLedResponse {
    return wfInst.SetLeds(sourceUri, LED_BREATHE, LedInfo{Count: -1, Colors: LedColors{ Ring: color }})
}

func (wfInst *workflowInstance) Vibrate(sourceUri string, pattern []uint64) VibrateResponse {
    fmt.Println("vibrating with pattern", pattern)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := vibrateRequest{Type: "wf_api_vibrate_request", Id: id, Target: target, Pattern: pattern}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := VibrateResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) getDeviceInfo(sourceUri string, query DeviceInfoQuery, refresh bool) GetDeviceInfoResponse {
    fmt.Println("getting device info with query", query, "refresh", refresh)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := getDeviceInfoRequest{Type: "wf_api_get_device_info_request", Id: id, Target: target, Query: query, Refresh: refresh}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := GetDeviceInfoResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
//     fmt.Println("device info response", res)
    return res
}

func (wfInst *workflowInstance) GetDeviceName(sourceUri string, refresh bool) string {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_NAME, refresh)
    fmt.Println("device info name", resp.Name)
    return resp.Name
}

func (wfInst *workflowInstance) GetDeviceId(sourceUri string, refresh bool) string {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_ID, refresh)
    fmt.Println("device info id", resp.Id)
    return resp.Id
}

func (wfInst *workflowInstance) GetDeviceAddress(sourceUri string, refresh bool) string {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_ADDRESS, refresh)
    fmt.Println("device info address", resp.Address)
    return resp.Address
}

func (wfInst *workflowInstance) GetDeviceLatLong(sourceUri string, refresh bool) []float64 {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_LATLONG, refresh)
    fmt.Println("device info latlong", resp.LatLong)
    return resp.LatLong
}

func (wfInst *workflowInstance) GetDeviceIndoorLocation(sourceUri string, refresh bool) string {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_INDOOR_LOCATION, refresh)
    fmt.Println("device info indoor location", resp.IndoorLocation)
    return resp.IndoorLocation
}

func (wfInst *workflowInstance) GetDeviceBattery(sourceUri string, refresh bool) uint64 {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_BATTERY, refresh)
    fmt.Println("device info battery", resp.Battery)
    return resp.Battery
}

func (wfInst *workflowInstance) GetDeviceType(sourceUri string, refresh bool) string {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_TYPE, refresh)
    fmt.Println("device info type", resp.Type)
    return resp.Type
}

func (wfInst *workflowInstance) GetDeviceUsername(sourceUri string, refresh bool) string {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_USERNAME, refresh)
    fmt.Println("device info username", resp.Username)
    return resp.Username
}

func (wfInst *workflowInstance) GetDeviceLocationEnabled(sourceUri string, refresh bool) bool {
    resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_LOCATION_ENABLED, refresh)
    fmt.Println("device info location enabled", resp.LocationEnabled)
    return resp.LocationEnabled
}

func (wfInst *workflowInstance) setDeviceInfo(sourceUri string, field SetDeviceInfoType, value string) SetDeviceInfoResponse {
    fmt.Println("setting device info field", field, "to", value)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := setDeviceInfoRequest{Type: "wf_api_set_device_info_request", Id: id, Target: target, Field: field, Value: value}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SetDeviceInfoResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) SetDeviceName(sourceUri string, name string) SetDeviceInfoResponse {
    return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_LABEL, name)
}

// func (wfInst *workflowInstance) SetDeviceChannel(sourceUri string, channel string) SetDeviceInfoResponse {
//     return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_CHANNEL, channel)
// }

func (wfInst *workflowInstance) EnableLocation(sourceUri string) SetDeviceInfoResponse {
    return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_LOCATION_ENABLED, "true")
}

func (wfInst *workflowInstance) DisableLocation(sourceUri string) SetDeviceInfoResponse {
    return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_LOCATION_ENABLED, "false")
}

func (wfInst *workflowInstance) SetUserProfile(sourceUri string, username string, force bool) SetUserProfileResponse {
    fmt.Println("setting user profile to", username, "force", force)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := setUserProfileRequest{Type: "wf_api_set_user_profile_request", Id: id, Target: target, Username: username, Force: force}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SetUserProfileResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) SetChannel(sourceUri string, channelName string, suppressTTS bool, disableHomeChannel bool) SetChannelResponse {
    fmt.Println("setting channel", channelName, "suppressTTS", suppressTTS, "disableHomeChannel", disableHomeChannel)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := setChannelRequest{Type: "wf_api_set_channel_request", Id: id, Target: target, ChannelName: channelName, SuppressTTS: suppressTTS, DisableHomeChannel: disableHomeChannel}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SetChannelResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) SetDeviceMode(sourceUri string, mode DeviceMode) SetDeviceModeResponse {
    fmt.Println("setting device mode", mode)
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := setDeviceModeRequest{Type: "wf_api_set_device_mode_request", Id: id, Target: target, Mode: mode}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := SetDeviceModeResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) RestartDevice(sourceUri string) DevicePowerOffResponse {
    fmt.Println("restarting device")
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := devicePowerOffRequest{Type: "wf_api_device_power_off_request", Id: id, Target: target, Restart: true}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := DevicePowerOffResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) PowerDownDevice(sourceUri string) DevicePowerOffResponse {
    fmt.Println("powering down device")
    id := makeId()
    target := makeTargetMap(sourceUri)
    req := devicePowerOffRequest{Type: "wf_api_device_power_off_request", Id: id, Target: target, Restart: false}
    call := wfInst.sendAndReceiveRequest(req, id)
    res := DevicePowerOffResponse{}
    json.Unmarshal(call.EventWrapper.Msg, &res)
    return res
}

func (wfInst *workflowInstance) Terminate() {
    fmt.Println("terminating")
    id := makeId()
    req := terminateRequest{Type: "wf_api_terminate_request", Id: id}
    wfInst.sendRequest(req)
}
