// Copyright © 2022 Relay Inc.

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var workflowMap map[string]func(api RelayApi) = make(map[string]func(api RelayApi))

type RelayApi interface { // this is interface of your custom workflow, you implement this, then we call it and pass in the ws
	// assigning callbacks
	OnStart(fn func(startEvent StartEvent))
	OnInteractionLifecycle(fn func(interactionLifecycleEvent InteractionLifecycleEvent))
	OnPrompt(fn func(promptEvent PromptEvent)) // seperate into start and stop?
	OnTimerFired(fn func(timerFiredEvent TimerFiredEvent))
	OnButton(fn func(buttonEvent ButtonEvent))
	OnTimer(fn func(timerEvent TimerEvent))
	OnSpeech(fn func(speechEvent SpeechEvent))
	OnStop(fn func(stopEvent StopEvent))
	OnNotification(fn func(notificationEvent NotificationEvent))
	OnProgress(fn func(progressEvent ProgressEvent))
	OnPlayInboxMessages(fn func(playInboxMessagesEvent PlayInboxMessagesEvent))
	OnCallConnected(fn func(callConnectedEvent CallConnectedEvent))
	OnCallDisconnected(fn func(callDisconnectedEvent CallDisconnectedEvent))
	OnCallFailed(fn func(callFailedEvent CallFailedEvent))
	OnCallReceived(fn func(callReceivedEvent CallReceivedEvent))
	OnCallRinging(fn func(callRingingEvent CallRingingEvent))
	OnCallStartRequest(fn func(callStartEvent CallStartEvent))
	OnCallProgressing(fn func(callProgressingEvent CallProgressingEvent))
	OnSms(fn func(smsEvent SmsEvent))
	OnIncident(fn func(incidentEvent IncidentEvent))
	OnResume(fn func(resumeEvent ResumeEvent))

	// api
	GetSourceUri(startEvent StartEvent) string
	StartInteraction(sourceUri string, name string) StartInteractionResponse
	EndInteraction(sourceUri string) EndInteractionResponse
	SetTimer(timerType TimerType, name string, timeout uint64, timeoutType TimeoutType) SetTimerResponse
	ClearTimer(name string) ClearTimerResponse
	StartTimer(timeout int) StartTimerResponse // need to test timers
	CreateIncident(originator string, itype string) CreateIncidentResponse
	ResolveIncident(incidentId string, reason string) ResolveIncidentResponse
	Say(sourceUri string, text string, lang Language) SayResponse
	Alert(target string, originator string, name string, text string, pushOptions NotificationOptions) SendNotificationResponse
	CancelAlert(target string, name string) SendNotificationResponse
	SayAndWait(sourceUri string, text string, lang Language) SayResponse
	Listen(sourceUri string, phrases []string, transcribe bool, alt_lang Language, timeout int) string
	Translate(sourceUri string, text string, from Language, to Language) string
	LogMessage(message string, category string) LogAnalyticsEventResponse
	LogUserMessage(message string, sourceUri string, category string) LogAnalyticsEventResponse
	DebugLog(message string) DebugLogResponse
	SetVar(name string, value string) SetVarResponse
	UnsetVar(name string) UnsetVarResponse
	GetVar(name string, defaultValue string) string
	GetNumberVar(name string, defaultValue int) int
	Play(sourceUri string, filename string) string
	PlayAndWait(sourceUri string, filename string) string
	StopPlayback(sourceUri string, ids []string) StopPlaybackResponse
	GetUnreadInboxSize(sourceUri string) int
	PlayUnreadInboxMessages(sourceUri string) PlayInboxMessagesResponse
	SwitchLedOn(sourceUri string, ledIndex int, color string) SetLedResponse
	SwitchAllLedOn(sourceUri string, color string) SetLedResponse
	SwitchAllLedOff(sourceUri string) SetLedResponse
	Rainbow(sourceUri string, rotations int64) SetLedResponse
	Rotate(sourceUri string, color string, rotations int64) SetLedResponse
	Flash(sourceUri string, color string, count int64) SetLedResponse
	Breathe(sourceUri string, color string, count int64) SetLedResponse
	Vibrate(sourceUri string, pattern []int64) VibrateResponse
	Broadcast(target string, originator string, name string, text string, pushOptions NotificationOptions) SendNotificationResponse
	CancelBroadcast(target string, name string) SendNotificationResponse
	GetDeviceName(sourceUri string, refresh bool) string
	GetDeviceId(sourceUri string, refresh bool) string
	GetDeviceAddress(sourceUri string, refresh bool) string
	GetDeviceLocation(sourceUri string, refresh bool) string
	GetDeviceLatLong(sourceUri string, refresh bool) []float64
	IsGroupMember(groupNameUri string, potentialMemberUri string) bool
	GetGroupMembers(groupUri string) []string
	GetDeviceCoordinates(sourceUri string, refresh bool) []float64
	GetDeviceIndoorLocation(sourceUri string, refresh bool) string
	GetDeviceBattery(sourceUri string, refresh bool) uint64
	GetDeviceType(sourceUri string, refresh bool) string
	GetUserProfile(sourceUri string, refresh bool) string
	GetDeviceLocationEnabled(sourceUri string, refresh bool) bool
	SetDeviceName(sourceUri string, name string) SetDeviceInfoResponse
	EnableHomeChannel(sourceUri string) SetHomeChannelStateResponse
	DisableHomeChannel(sourceUri string) SetHomeChannelStateResponse
	// SetDeviceChannel(sourceUri string, channel string) SetDeviceInfoResponse
	EnableLocation(sourceUri string) SetDeviceInfoResponse
	DisableLocation(sourceUri string) SetDeviceInfoResponse
	SetUserProfile(sourceUri string, username string, force bool) SetUserProfileResponse
	SetChannel(sourceUri string, channelName string, suppressTTS bool, disableHomeChannel bool) SetChannelResponse
	// SetDeviceMode(sourceUri string, mode DeviceMode) SetDeviceModeResponse
	// RestartDevice(sourceUri string) DevicePowerOffResponse
	// PowerDownDevice(sourceUri string) DevicePowerOffResponse
	PlaceCall(targetUri string, uri string) PlaceCallResponse
	AnswerCall(sourceUri string, callId string) AnswerResponse
	HangupCall(targetUri string, callId string) HangupCallResponse
	Terminate()
}

// This struct implements RelayApi below
type workflowInstance struct {
	WebsocketConnection *websocket.Conn
	Mutex               sync.Mutex       // no initialization, zero value is unlocked mutex. this must not be copied, always pass workflowInstance by pointer
	Pending             map[string]*Call // map of request ids to the call struct for response pairing
	WorkflowFn          func(api RelayApi)

	EventChannel chan EventWrapper
	StopReason   string

	// stores callback functions for each event type
	OnStartHandler                func(startEvent StartEvent)
	OnInteractionLifecycleHandler func(interactionLifecycleEvent InteractionLifecycleEvent)
	OnPromptHandler               func(promptEvent PromptEvent)
	OnButtonHandler               func(buttonEvent ButtonEvent)
	OnTimerFiredHandler           func(timerFiredEvent TimerFiredEvent)
	OnTimerHandler                func(timerEvent TimerEvent)
	OnSpeechHandler               func(speechEvent SpeechEvent)
	OnStopHandler                 func(stopEvent StopEvent)
	OnNotificationHandler         func(notificationEvent NotificationEvent)
	OnProgressHandler             func(progressEvent ProgressEvent)
	OnPlayInboxMessagesHandler    func(playInboxMessagesEvent PlayInboxMessagesEvent)
	OnCallConnectedHandler        func(callConnectedEvent CallConnectedEvent)
	OnCallDisconnectedHandler     func(callDisconnected CallDisconnectedEvent)
	OnCallFailedHandler           func(callFailedEvent CallFailedEvent)
	OnCallReceivedHandler         func(callReceivedEvent CallReceivedEvent)
	OnCallRingingHandler          func(callRingingEvent CallRingingEvent)
	OnCallStartRequestHandler     func(callStartEvent CallStartEvent)
	OnCallProgressingHandler      func(callProgressingEvent CallProgressingEvent)
	OnSmsHandler                  func(smsEvent SmsEvent)
	OnIncidentHandler             func(incidentEvent IncidentEvent)
	OnResumeHandler               func(resumeEvent ResumeEvent)
}

// Call represents an active request
type Call struct {
	Req          interface{}
	Res          interface{}
	EventWrapper EventWrapper
	Done         chan bool
	Error        error
}

type EventWrapper struct {
	ParsedMsg map[string]interface{}
	Msg       []byte
	EventName Event
}

// Callback Handlers

// A decorator for a handler for the START event (workflow is starting).
func (wfInst *workflowInstance) OnStart(fn func(startEvent StartEvent)) {
	// store the func that was passed in as a callback in a slice, then whenever the websocket sends us a matching event, call the callback
	wfInst.OnStartHandler = fn
}

// A decorator for a handler method for the INTERACTION_LIFECYCLE event (an interaction
// is starting, resuming, or ending).
func (wfInst *workflowInstance) OnInteractionLifecycle(fn func(interactionLifecycleEvent InteractionLifecycleEvent)) {
	wfInst.OnInteractionLifecycleHandler = fn
}

// A decorator for a handler method for the PROMPT event (text-to-speech is streaming in).
func (wfInst *workflowInstance) OnPrompt(fn func(promptEvent PromptEvent)) {
	wfInst.OnPromptHandler = fn
}

// A decorator for a handler method for the BUTTON event (the Talk button was pressed).
func (wfInst *workflowInstance) OnButton(fn func(buttonEvent ButtonEvent)) {
	wfInst.OnButtonHandler = fn
}

// A decorator for a handler method for the TIMER_FIRED event (a named timer fired).
func (wfInst *workflowInstance) OnTimerFired(fn func(timerFiredEvent TimerFiredEvent)) {
	wfInst.OnTimerFiredHandler = fn
}

// A decorator for a handler method for the TIMER event (the unnamed timer fired).
func (wfInst *workflowInstance) OnTimer(fn func(timerEvent TimerEvent)) {
	wfInst.OnTimerHandler = fn
}

// A decorator for a handler method for the SPEECH event (the listen() function is running).
func (wfInst *workflowInstance) OnSpeech(fn func(speechEvent SpeechEvent)) {
	wfInst.OnSpeechHandler = fn
}

// A decorator for a handler method for the STOP event (workflow is stopping).
func (wfInst *workflowInstance) OnStop(fn func(stopEvent StopEvent)) {
	wfInst.OnStopHandler = fn
}

// A decorator for a handler method for the NOTIFICATION event (a broadcast or alert was sent).
func (wfInst *workflowInstance) OnNotification(fn func(notificationEvent NotificationEvent)) {
	wfInst.OnNotificationHandler = fn
}

// A decorator for a handler method for the PROGRESS event (a long running action is being
// performed across a large number of devices, may get called multiple times).
func (wfInst *workflowInstance) OnProgress(fn func(progressEvent ProgressEvent)) {
	wfInst.OnProgressHandler = fn
}

// A decorator for a handler method for the PLAY_INBOX_MESSAGE event (a missed message
// is being played).
func (wfInst *workflowInstance) OnPlayInboxMessages(fn func(playInboxMessagesEvent PlayInboxMessagesEvent)) {
	wfInst.OnPlayInboxMessagesHandler = fn
}

// A decorator for a handler method for the CALL_CONNECTED event.
// A call attempt that was ringing, progressing, or incoming is now fully
// connected. This event can occur on both the caller and the callee.
func (wfInst *workflowInstance) OnCallConnected(fn func(callConnectedEvent CallConnectedEvent)) {
	wfInst.OnCallConnectedHandler = fn
}

// A decorator for a handler method for the CALL_DISCONNECTED event.
// A call that was once connected has become disconnected. This event can
// occur on both the caller and the callee.
func (wfInst *workflowInstance) OnCallDisconnected(fn func(callDisconnected CallDisconnectedEvent)) {
	wfInst.OnCallDisconnectedHandler = fn
}

// A decorator for a handler method for the CALL_FAILED event.
// A call failed to get connected. This event can occur on both the caller
// and the callee.
func (wfInst *workflowInstance) OnCallFailed(fn func(callFailedEvent CallFailedEvent)) {
	wfInst.OnCallFailedHandler = fn
}

// A decorator for a handler method for the CALL_RECEIVED event.
// The device is receiving an inbound call request. This event can occur
// on the callee.
func (wfInst *workflowInstance) OnCallReceived(fn func(callReceivedEvent CallReceivedEvent)) {
	wfInst.OnCallReceivedHandler = fn
}

// A decorator for a handler method for the CALL_RINGING event.
// The device we called is ringing. We are waiting for them to answer.
// This event can occur on the caller.
func (wfInst *workflowInstance) OnCallRinging(fn func(callRingingEvent CallRingingEvent)) {
	wfInst.OnCallRingingHandler = fn
}

// A decorator for a handler method for the CALL_START_REQUEST event.
// There is a request to make an outbound call. This event can occur on
// the caller after using the "Call X" voice command on the Assistant.
func (wfInst *workflowInstance) OnCallStartRequest(fn func(callStartEvent CallStartEvent)) {
	wfInst.OnCallStartRequestHandler = fn
}

// A decorator for a handler method for the CALL_PROGRESSING event.
// The device we called is making progress on getting connected. This may
// be interspersed with on_call_ringing. This event can occur on the caller.
func (wfInst *workflowInstance) OnCallProgressing(fn func(callProgressingEvent CallProgressingEvent)) {
	wfInst.OnCallProgressingHandler = fn
}

// A decorator for a handler method for the SMS event (TBD).
func (wfInst *workflowInstance) OnSms(fn func(smsEvent SmsEvent)) {
	wfInst.OnSmsHandler = fn
}

// A decorator for a handler method for the INCIDENT event (an incident has been created).
func (wfInst *workflowInstance) OnIncident(fn func(incidentEvent IncidentEvent)) {
	wfInst.OnIncidentHandler = fn
}

// A decorator for a handler method for the RESUME event (TBD).
func (wfInst *workflowInstance) OnResume(fn func(resumeEvent ResumeEvent)) {
	wfInst.OnResumeHandler = fn
}

// API functions

// Helper method for parsing out the source URN from a start event trigger.
func (wfInst *workflowInstance) GetSourceUri(startEvent StartEvent) string {
	return startEvent.Trigger.Args.SourceUri
}

// Starts an interaction with the user.  Triggers an INTERACTION_STARTED event
// and allows the user to interact with the device via functions that require an
// interaction URN. Returns a StartInteractionResponse.
func (wfInst *workflowInstance) StartInteraction(sourceUri string, name string) StartInteractionResponse {
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := startInteractionRequest{Type: "wf_api_start_interaction_request", Id: id, Targets: target, Name: name}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := StartInteractionResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Ends an interaction with the user.  Triggers an INTERACTION_ENDED event to signify
// that the user is done interacting with the device.  Returns an EndInteractionResponse.
func (wfInst *workflowInstance) EndInteraction(sourceUri string) EndInteractionResponse {
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := endInteractionRequest{Type: "wf_api_end_interaction_request", Id: id, Targets: target}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := EndInteractionResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Serves as a named timer that can be either interval or timeout.  Allows you to specify
// the unit of time. Returns a SetTimerResponse.
func (wfInst *workflowInstance) SetTimer(timerType TimerType, name string, timeout uint64, timeoutType TimeoutType) SetTimerResponse {
	id := makeId()
	req := setTimerRequest{Type: "wf_api_set_timer_request", Id: id, TimerType: timerType, Name: name, Timeout: timeout, TimeoutType: timeoutType}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetTimerResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Clears the specified timer. Returns a ClearTimerResponse.
func (wfInst *workflowInstance) ClearTimer(name string) ClearTimerResponse {
	id := makeId()
	req := clearTimerRequest{Type: "wf_api_clear_timer_request", Id: id, Name: name}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := ClearTimerResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Starts an unnamed timer, meaning this will be the only timer on your device.
// The timer will fire when it reaches the value of the 'timeout' parameter. Returns
// a StartTimerResponse.
func (wfInst *workflowInstance) StartTimer(timeout int) StartTimerResponse {
	id := makeId()
	req := startTimerRequest{Type: "wf_api_start_timer_request", Id: id, Timeout: timeout}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := StartTimerResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Stops an unnamed timer.  Returns a StopTimerResponse.
func (wfInst *workflowInstance) StopTimer() StopTimerResponse {
	id := makeId()
	req := stopTimerRequest{Type: "wf_api_stop_timer_request", Id: id}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := StopTimerResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Creates an incident that will alert the Relay Dash. Returns a CreateIncidentResponse.
func (wfInst *workflowInstance) CreateIncident(originator string, itype string) CreateIncidentResponse {
	id := makeId()
	req := createIncidentRequest{Type: "wf_api_create_incident_request", Id: id, IncidentType: itype, OriginatorUri: originator}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := CreateIncidentResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Resolved an incident that was created. Returns a ResolveIncidentResponse.
func (wfInst *workflowInstance) ResolveIncident(incidentId string, reason string) ResolveIncidentResponse {
	id := makeId()
	req := resolveIncidentRequest{Type: "wf_api_resolve_incident_request", Id: id, IncidentId: incidentId, Reason: reason}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := ResolveIncidentResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Utilizes text to speech capabilities to make the device 'speak' to the user. Returns a SayResponse.
func (wfInst *workflowInstance) Say(sourceUri string, text string, lang Language) SayResponse {
	if lang == "" {
		lang = ENGLISH
	}
	log.Debug("saying ", text, " to ", sourceUri, " with lang ", lang)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := sayRequest{Type: "wf_api_say_request", Id: id, Target: target, Text: text, Lang: lang}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SayResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Utilizes text to speech capabilities to make the device 'speak' to the user.
// Waits until the text is fully played out on the device before continuing. Returns a SayResponse.
func (wfInst *workflowInstance) SayAndWait(sourceUri string, text string, lang Language) SayResponse {
	if lang == "" {
		lang = ENGLISH
	}
	log.Debug("saying ", text, " to ", sourceUri, " with lang ", lang)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := sayRequest{Type: "wf_api_say_request", Id: id, Target: target, Text: text, Lang: lang}
	call := wfInst.sendAndReceiveRequestWait(req, id)
	res := SayResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Listens for the user to speak into the device.  Utilizes speech to text functionality to interact
// with the user. Returns the text that the device parsed from the speech as a string.
func (wfInst *workflowInstance) Listen(sourceUri string, phrases []string, transcribe bool, alt_lang Language, timeout int) string {
	log.Debug("listening ")
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := listenRequest{Type: "wf_api_listen_request", Id: id, Target: target, ReqestId: "request1", Phrases: phrases, Transcribe: transcribe, Timeout: timeout, AltLang: string(alt_lang)}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SpeechEvent{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res.Text
}

// Translates text from one language to another. Returns the translated text in the specified language as a string.
func (wfInst *workflowInstance) Translate(sourceUri string, text string, from Language, to Language) string {
	log.Debug("translating ", text)
	id := makeId()
	req := translateRequest{Type: "wf_api_translate_request", Id: id, Text: text, FromLang: from, ToLang: to}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := TranslateResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res.Text
}

// Log an analytics event from a workflow with the specified content and
// under a specified category. This does not log the device who
// triggered the workflow that called this function. Returns a LogAnalyticsEventResponse.
func (wfInst *workflowInstance) LogMessage(message string, category string) LogAnalyticsEventResponse {
	log.Debug("logging analytic event with the message ", message)
	id := makeId()
	req := logAnalyticsEventRequest{Type: "wf_api_log_analytics_event_request", Id: id, Content: message, ContentType: "default", Category: category}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := LogAnalyticsEventResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Log an analytic event from a workflow with the specified content and
// under a specified category.  This includes the device who triggered the workflow
// that called this function. Returns a LogAnalyticsEventResponse.
func (wfInst *workflowInstance) LogUserMessage(message string, sourceUri string, category string) LogAnalyticsEventResponse {
	log.Debug("logging analytic event with the message ", message)
	id := makeId()
	req := logAnalyticsEventRequest{Type: "wf_api_log_analytics_event_request", Id: id, Content: message, ContentType: "default", Category: category, DeviceUri: sourceUri}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := LogAnalyticsEventResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Log a debug message that is visible with the CLI command
// `relay workflow debug`. This is helpful for debugging workflows.
// Returns a DebugLogResponse.
func (wfInst *workflowInstance) DebugLog(message string) DebugLogResponse {
	log.Debug("logging debug message ", message)
	id := makeId()
	req := debugLogRequest{Type: "wf_api_debug_log_request", Id: id, Content: message}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := DebugLogResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Sets a variable with the corresponding name and value. Scope of
// the variable is from start to end of a workflow.  Note that you
// can only set values of type string. Returns a SetVarResponse.
func (wfInst *workflowInstance) SetVar(name string, value string) SetVarResponse {
	log.Debug("setting variable with name ", name, " and value ", value)
	id := makeId()
	req := setVarRequest{Type: "wf_api_set_var_request", Id: id, Name: name, Value: value}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetVarResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Unsets the value of a variable. Returns an UnsetVarResponse.
func (wfInst *workflowInstance) UnsetVar(name string) UnsetVarResponse {
	log.Debug("unsetting variable with name ", name)
	id := makeId()
	req := unsetVarRequest{Type: "wf_api_unset_var_request", Id: id, Name: name}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := UnsetVarResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Retrieves a variable that was set either during workflow registration
// or through the set_var() function.  The variable can be retrieved anywhere
// within the workflow, but is erased after the workflow terminates. Returns the
// requested variable's value as a string.
func (wfInst *workflowInstance) GetVar(name string, defaultValue string) string {
	log.Debug("getting variable with name ", name, " and default value ", defaultValue)
	id := makeId()
	req := getVarRequest{Type: "wf_api_get_var_request", Id: id, Name: name}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := GetVarResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	if res.Value != "" {
		return res.Value
	}
	return defaultValue
}

// Retrieves a variable that was set either during workflow registration
// or through the set_var() function of type integer.  The variable can be retrieved anywhere
// within the workflow, but is erased after the workflow terminates. Returns the requested
// variable's value as an integer.
func (wfInst *workflowInstance) GetNumberVar(name string, defaultValue int) int {
	numVar, err := strconv.Atoi(wfInst.GetVar(name, strconv.FormatInt(int64(defaultValue), 10)))
	log.Error(err)
	return numVar
}

// Plays a custom audio file that was uploaded by the user. Returns the correlation ID retrieved
// from the PlayResponse as a string.
func (wfInst *workflowInstance) Play(sourceUri string, filename string) string {
	log.Debug("playing file ", filename, " to ", sourceUri)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := playRequest{Type: "wf_api_play_request", Id: id, Target: target, Filename: filename}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := PlayResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res.CorrelationId
}

// Plays a custom audio file that was uploaded by the user.
// Waits until the audio file has finished playing before continuing through
// the workflow. Returns the correlation ID retrieved from the PlayResponse as a string.
func (wfInst *workflowInstance) PlayAndWait(sourceUri string, filename string) string {
	log.Debug("playing file ", filename, " to ", sourceUri)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := playRequest{Type: "wf_api_play_request", Id: id, Target: target, Filename: filename}
	call := wfInst.sendAndReceiveRequestWait(req, id)
	res := PlayResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res.CorrelationId
}

// Stops a playback request on the device. Returns the StopPlaybackResponse.
func (wfInst *workflowInstance) StopPlayback(sourceUri string, ids []string) StopPlaybackResponse {
	log.Debug("stopping playback for ", ids)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := stopPlaybackRequest{Type: "wf_api_stop_playback_request", Id: id, Target: target, Ids: ids}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := StopPlaybackResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Retrieves the number of messages in device's inbox. Returns the number
// of unread messages in the device's inbox as an integer.
func (wfInst *workflowInstance) GetUnreadInboxSize(sourceUri string) int {
	log.Debug("playing unread inbox messages for ", sourceUri)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := inboxCountRequest{Type: "wf_api_inbox_count_request", Id: id, Target: target}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := InboxCountResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	count, err := strconv.Atoi(res.Count)
	fmt.Println("error", err)
	return count
}

// Play a targeted device's inbox messages. Returns the PlayInboxMessagesResponse.
func (wfInst *workflowInstance) PlayUnreadInboxMessages(sourceUri string) PlayInboxMessagesResponse {
	log.Debug("playing unread inbox messages for ", sourceUri)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := playInboxMessagesRequest{Type: "wf_api_play_inbox_messages_request", Id: id, Target: target}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := PlayInboxMessagesResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

func (wfInst *workflowInstance) setHomeChannelState(sourceUri string, enabled bool) SetHomeChannelStateResponse {
	log.Debug("setting home channel for ", sourceUri, " with state ", enabled)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := setHomeChannelStateRequest{Type: "wf_api_set_home_channel_state_request", Id: id, Target: target, Enabled: enabled}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetHomeChannelStateResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Enables the home channel on the device. Returns the SetHomeChannelStateResponse.
func (wfInst *workflowInstance) EnableHomeChannel(sourceUri string) SetHomeChannelStateResponse {
	return wfInst.setHomeChannelState(sourceUri, true)
}

// Disables the home channel on the device. Returns the SetHomeChannelStateResponse.
func (wfInst *workflowInstance) DisableHomeChannel(sourceUri string) SetHomeChannelStateResponse {
	return wfInst.setHomeChannelState(sourceUri, false)
}

func (wfInst *workflowInstance) setLeds(sourceUri string, effect LedEffect, args LedInfo) SetLedResponse {
	log.Debug("setting leds ", effect, " with args ", args)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := setLedRequest{Type: "wf_api_set_led_request", Id: id, Target: target, Effect: effect, Args: args}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetLedResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Switches on an LED at a particules index to a specified color. Returns a SetLedResponse.
func (wfInst *workflowInstance) SwitchLedOn(sourceUri string, led int, color string) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_STATIC, LedInfo{Colors: setLedColors(strconv.FormatInt(int64(led), 10), color)})
}

// Switches all the LEDs on a device on to a specified color. Returns a SetLedResponse.
func (wfInst *workflowInstance) SwitchAllLedOn(sourceUri string, color string) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_STATIC, LedInfo{Colors: LedColors{Ring: color}})
}

// Swithes all of the LEDs on a device off. Returns a SetLedResponse.
func (wfInst *workflowInstance) SwitchAllLedOff(sourceUri string) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_OFF, LedInfo{})
}

// Switches all the LEDs on to a configured rainbow pattern and rotates the rainbow
// a specified number of times. Returns a SetLedResponse.
func (wfInst *workflowInstance) Rainbow(sourceUri string, rotations int64) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_RAINBOW, LedInfo{Rotations: rotations})
}

// Switches all of the LEDs on a device to a certain color and rotates them a specified number
// of times. Returns a SetLedResponse.
func (wfInst *workflowInstance) Rotate(sourceUri string, color string, rotations int64) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_ROTATE, LedInfo{Rotations: rotations, Colors: LedColors{Led1: color}})
}

// Switches all of the LEDs on a device to a certain color and flashes them
// a specified number of times. Returns a SetLedResponse.
func (wfInst *workflowInstance) Flash(sourceUri string, color string, count int64) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_FLASH, LedInfo{Count: count, Colors: LedColors{Ring: color}})
}

// Switches all of the LEDs on a device to a certain color and creates a 'breathing' effect,
// where the LEDs will slowly light up a specified number of times. Returns a SetLedResponse.
func (wfInst *workflowInstance) Breathe(sourceUri string, color string, count int64) SetLedResponse {
	return wfInst.setLeds(sourceUri, LED_BREATHE, LedInfo{Count: count, Colors: LedColors{Ring: color}})
}

// Makes the device vibrate in a particular pattern.  You can specify
// how many vibrations you would like, the duration of each vibration in
// milliseconds, and how long you would like the pauses between each vibration to last
// in milliseconds. Returns a VibrateResponse.
func (wfInst *workflowInstance) Vibrate(sourceUri string, pattern []int64) VibrateResponse {
	log.Debug("vibrating with pattern ", pattern)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := vibrateRequest{Type: "wf_api_vibrate_request", Id: id, Target: target, Pattern: pattern}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := VibrateResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

func (wfInst *workflowInstance) sendNotification(target string, originator string, itype string, name string, text string, pushOptions NotificationOptions) SendNotificationResponse {
	log.Debug("sending a notification of type ", itype)
	id := makeId()
	targetMap := makeTargetMap(target)
	req := sendNotificationRequest{Type: "wf_api_notification_request", Id: id, Target: targetMap, Originator: originator, IType: itype, Name: name, Text: text, ITarget: targetMap, PushOptions: pushOptions}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SendNotificationResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Sends out a broadcasted message to a group of devices.  The message is played out on
// all devices, as well as sent to the Relay Dash. Returns a SendNotificationResponse.
func (wfInst *workflowInstance) Broadcast(target string, originator string, name string, text string, pushOptions NotificationOptions) SendNotificationResponse {
	return wfInst.sendNotification(target, originator, "broadcast", name, text, pushOptions)
}

// Cancels the broadcsat that was sent to a group of devices. Returns a SendNotificationResponse.
func (wfInst *workflowInstance) CancelBroadcast(target string, name string) SendNotificationResponse {
	var pushOptions NotificationOptions
	return wfInst.sendNotification(target, "", "cancel", name, "", pushOptions)
}

// Sends out an alert to the specified group of devices and the Relay Dash. Returns a SendNotificationResponse.
func (wfInst *workflowInstance) Alert(target string, originator string, name string, text string, pushOptions NotificationOptions) SendNotificationResponse {
	return wfInst.sendNotification(target, originator, "alert", name, text, pushOptions)
}

// Cancels an alert that was sent to a group of devices.  Particularly useful if you would like to cancel the alert
// on all devices after one device has acknowledged the alert. Returns a SendNotificationResponse.
func (wfInst *workflowInstance) CancelAlert(target string, name string) SendNotificationResponse {
	var pushOptions NotificationOptions
	return wfInst.sendNotification(target, "", "cancel", name, "", pushOptions)
}

func (wfInst *workflowInstance) getDeviceInfo(sourceUri string, query DeviceInfoQuery, refresh bool) GetDeviceInfoResponse {
	log.Debug("getting device info with query ", query, " refresh ", refresh)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := getDeviceInfoRequest{Type: "wf_api_get_device_info_request", Id: id, Target: target, Query: query, Refresh: refresh}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := GetDeviceInfoResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Returns the name of a targeted device as a string.
func (wfInst *workflowInstance) GetDeviceName(sourceUri string, refresh bool) string {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_NAME, refresh)
	log.Debug("device info name ", resp.Name)
	return resp.Name
}

// Returns the ID of the targeted device as a string.
func (wfInst *workflowInstance) GetDeviceId(sourceUri string, refresh bool) string {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_ID, refresh)
	log.Debug("device info id ", resp.Id)
	return resp.Id
}

// Returns the location of a targeted device as a string.
func (wfInst *workflowInstance) GetDeviceLocation(sourceUri string, refresh bool) string {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_ADDRESS, refresh)
	log.Debug("device info address ", resp.Address)
	return resp.Address
}

// Returns the address of a targeted device as a string.
func (wfInst *workflowInstance) GetDeviceAddress(sourceUri string, refresh bool) string {
	return wfInst.GetDeviceLocation(sourceUri, refresh)
}

// Retrieves the coordinates of the device's location. Returns a float64 array containing the coordinates of
// the device.
func (wfInst *workflowInstance) GetDeviceCoordinates(sourceUri string, refresh bool) []float64 {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_LATLONG, refresh)
	log.Debug("device info latlong ", resp.LatLong)
	return resp.LatLong
}

// Returns the latitude and longitude coordinates of a targeted device. Returns a float64 array containing thecoordinates
// of the device.
func (wfInst *workflowInstance) GetDeviceLatLong(sourceUri string, refresh bool) []float64 {
	return wfInst.GetDeviceCoordinates(sourceUri, refresh)
}

// Returns the indoor location of a targeted device as a string.
func (wfInst *workflowInstance) GetDeviceIndoorLocation(sourceUri string, refresh bool) string {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_INDOOR_LOCATION, refresh)
	log.Debug("device info indoor location ", resp.IndoorLocation)
	return resp.IndoorLocation
}

// Returns the battery of a targeted device as a string.
func (wfInst *workflowInstance) GetDeviceBattery(sourceUri string, refresh bool) uint64 {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_BATTERY, refresh)
	log.Debug("device info battery ", resp.Battery)
	return resp.Battery
}

// Returns the device type of a targeted device, i.e. gen 2, gen 3, etc. as a string.
func (wfInst *workflowInstance) GetDeviceType(sourceUri string, refresh bool) string {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_TYPE, refresh)
	log.Debug("device info type ", resp.Type)
	return resp.Type
}

// Returns the user profile of a targeted device as a string.
func (wfInst *workflowInstance) GetUserProfile(sourceUri string, refresh bool) string {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_USERNAME, refresh)
	log.Debug("device info username ", resp.Username)
	return resp.Username
}

// Returns whether the location services on a device are enabled as a boolean.
func (wfInst *workflowInstance) GetDeviceLocationEnabled(sourceUri string, refresh bool) bool {
	resp := wfInst.getDeviceInfo(sourceUri, DEVICE_INFO_QUERY_LOCATION_ENABLED, refresh)
	log.Debug("device info location enabled ", resp.LocationEnabled)
	return resp.LocationEnabled
}

func (wfInst *workflowInstance) setDeviceInfo(sourceUri string, field SetDeviceInfoType, value string) SetDeviceInfoResponse {
	log.Debug("setting device info field ", field, " to ", value)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := setDeviceInfoRequest{Type: "wf_api_set_device_info_request", Id: id, Target: target, Field: field, Value: value}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetDeviceInfoResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Sets the name of a targeted device and updates it on the Relay Dash.
// The name remains updated until it is set again via a workflow or updated manually
// on the Relay Dash.  Returns a SendNotificationResponse.
func (wfInst *workflowInstance) SetDeviceName(sourceUri string, name string) SetDeviceInfoResponse {
	return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_LABEL, name)
}

// SetDeviceChannel is currently not supported

// func (wfInst *workflowInstance) SetDeviceChannel(sourceUri string, channel string) SetDeviceInfoResponse {
//     return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_CHANNEL, channel)
// }

// Enables location services on a device.  Location services will remain
// enabled until they are disabled on the Relay Dash or through a workflow.
func (wfInst *workflowInstance) EnableLocation(sourceUri string) SetDeviceInfoResponse {
	return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_LOCATION_ENABLED, "true")
}

// Disables location services on a device.  Location services will remain
// disabled until they are enabled on the Relay Dash or through a workflow. Returns a SendNotificationResponse.
func (wfInst *workflowInstance) DisableLocation(sourceUri string) SetDeviceInfoResponse {
	return wfInst.setDeviceInfo(sourceUri, SET_DEVICE_INFO_LOCATION_ENABLED, "false")
}

// Returns the members of a particular group as a string array.
func (wfInst *workflowInstance) GetGroupMembers(groupUri string) []string {
	log.Debug("retrieving members of ", groupUri)
	id := makeId()
	req := groupQueryRequest{Type: "wf_api_group_query_request", Id: id, GroupUri: groupUri, Query: "list_members"}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := GroupQueryResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res.MemberUris
}

// Checks whether a device is a member of a particular group. Returns true if the device is a member
// of the specified group, false otherwise.
func (wfInst *workflowInstance) IsGroupMember(groupNameUri string, potentialMemberUri string) bool {
	var groupName string = ParseGroupName(groupNameUri)
	var deviceName string = ParseDeviceName(potentialMemberUri)
	var groupUri string = GroupMember(groupName, deviceName)

	fmt.Println("retrieving whether ", deviceName, " is a part of group ", groupName)
	id := makeId()
	req := groupQueryRequest{Type: "wf_api_group_query_request", Id: id, GroupUri: groupUri, Query: "is_member"}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := GroupQueryResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res.IsMember
}

// Sets the profile of a user by updating the username. Returns a SetUserProfileResponse.
func (wfInst *workflowInstance) SetUserProfile(sourceUri string, username string, force bool) SetUserProfileResponse {
	log.Debug("setting user profile to ", username, " force ", force)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := setUserProfileRequest{Type: "wf_api_set_user_profile_request", Id: id, Target: target, Username: username, Force: force}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetUserProfileResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Sets the channel that a device is on.  This can be used to change the channel of a device during a workflow,
// where the channel will also be updated on the Relay Dash. Returns a SetChannelResponse.
func (wfInst *workflowInstance) SetChannel(sourceUri string, channelName string, suppressTTS bool, disableHomeChannel bool) SetChannelResponse {
	log.Debug("setting channel ", channelName, " suppressTTS ", suppressTTS, " disableHomeChannel ", disableHomeChannel)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := setChannelRequest{Type: "wf_api_set_channel_request", Id: id, Target: target, ChannelName: channelName, SuppressTTS: suppressTTS, DisableHomeChannel: disableHomeChannel}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := SetChannelResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// SetDeviceMode is currently not supported.

// func (wfInst *workflowInstance) SetDeviceMode(sourceUri string, mode DeviceMode) SetDeviceModeResponse {
//     log.Debug("setting device mode ", mode)
//     id := makeId()
//     target := makeTargetMap(sourceUri)
//     req := setDeviceModeRequest{Type: "wf_api_set_device_mode_request", Id: id, Target: target, Mode: mode}
//     call := wfInst.sendAndReceiveRequest(req, id)
//     res := SetDeviceModeResponse{}
//     json.Unmarshal(call.EventWrapper.Msg, &res)
//     return res
// }

// Restart/Powering down device is currently not supported

// func (wfInst *workflowInstance) RestartDevice(sourceUri string) DevicePowerOffResponse {
//     fmt.Println("restarting device")
//     id := makeId()
//     target := makeTargetMap(sourceUri)
//     req := devicePowerOffRequest{Type: "wf_api_device_power_off_request", Id: id, Target: target, Restart: true}
//     call := wfInst.sendAndReceiveRequest(req, id)
//     res := DevicePowerOffResponse{}
//     json.Unmarshal(call.EventWrapper.Msg, &res)
//     return res
// }

// func (wfInst *workflowInstance) PowerDownDevice(sourceUri string) DevicePowerOffResponse {
//     fmt.Println("powering down device")
//     id := makeId()
//     target := makeTargetMap(sourceUri)
//     req := devicePowerOffRequest{Type: "wf_api_device_power_off_request", Id: id, Target: target, Restart: false}
//     call := wfInst.sendAndReceiveRequest(req, id)
//     res := DevicePowerOffResponse{}
//     json.Unmarshal(call.EventWrapper.Msg, &res)
//     return res
// }

// Places a call to another device. Returns a PlaceCallResponse.
func (wfInst *workflowInstance) PlaceCall(targetUri string, uri string) PlaceCallResponse {
	log.Debug("placing call to ", targetUri, " with uri ", uri)
	id := makeId()
	target := makeTargetMap(targetUri)
	req := placeCallRequest{Type: "wf_api_call_request", Id: id, Target: target, Uri: uri}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := PlaceCallResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Answers a call on your device. Returns an AnswerResponse.
func (wfInst *workflowInstance) AnswerCall(sourceUri string, callId string) AnswerResponse {
	log.Debug("calling device with call id ", callId)
	id := makeId()
	target := makeTargetMap(sourceUri)
	req := answerRequest{Type: "wf_api_answer_request", Id: id, Target: target, CallId: callId}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := AnswerResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Ends a call on your device.  Note that target can only have one item. Returns a HangupCallResponse.
func (wfInst *workflowInstance) HangupCall(targetUri string, callId string) HangupCallResponse {
	log.Debug("hanging up call with ", callId, " and target uri ", targetUri)
	id := makeId()
	target := makeTargetMap(targetUri)
	req := hangupCallRequest{Type: "wf_api_hangup_request", Id: id, Target: target, CallId: callId}
	call := wfInst.sendAndReceiveRequest(req, id)
	res := HangupCallResponse{}
	json.Unmarshal(call.EventWrapper.Msg, &res)
	return res
}

// Terminates a workflow.  This method is usually called
// after your workflow has completed and you would like to end the
// workflow by calling end_interaction(), where you can then terminate
// the workflow.
func (wfInst *workflowInstance) Terminate() {
	log.Debug("terminating")
	id := makeId()
	req := terminateRequest{Type: "wf_api_terminate_request", Id: id}
	wfInst.sendRequest(req)
}

// Used only for TriggerWorkflow and FetchDevice
var serverHostname string = "all-main-pro-ibot.relaysvr.com"
var version string = "relay-sdk-go/2.0.0-pre"
var auth_hostname string = "auth.relaygo.com"

func updateAccessToken(refreshToken string, clientId string) string {
	grantUrl := "https://" + auth_hostname + "/oauth2/token"

	// Create the query params to be sent with the request, and encode the query params
	queryParams := url.Values{}
	queryParams.Add("grant_type", "refresh_token")
	queryParams.Add("refresh_token", refreshToken)
	queryParams.Add("client_id", clientId)

	var grantPayload = []byte(queryParams.Encode())

	// Create a new POST request with the URL and query parameters
	req, err := http.NewRequest("POST", grantUrl, bytes.NewBuffer(grantPayload))
	if err != nil {
		log.Error("Error creating HTTP request in updateAccessToken: ", err)
	}

	// Set the headers
	req.Header.Set("User-Agent", version)

	// Create a client and perform the request
	client := &http.Client{}
	client.Timeout = time.Second * 30
	res, err := client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request in updateAccessToken: ", err)
	}

	defer res.Body.Close()

	// Ensure a 200 status code
	if res.StatusCode != http.StatusOK {
		log.Error("Failed to retrieve access token with status code ", res.StatusCode)
	}

	// Convert the JSON response into a map and return the access token
	var accessTokenRes map[string]interface{}
	error := json.NewDecoder(res.Body).Decode(&accessTokenRes)
	if error != nil {
		log.Error("Error decoding response body in updateAccessToken: ", err)
	}
	return accessTokenRes["access_token"].(string)

}

// A convenience method for sending an HTTP trigger to the Relay server.
// This generally would be used in a third-party system to start a Relay
// workflow via an HTTP trigger and optionally pass data to it with
// action_args.
// If the access_token has expired and the request gets a 401 response,
// a new access_token will be automatically generated via the refresh_token,
// and the request will be resubmitted with the new access_token. Otherwise
// the refresh token won't be used.
// This method will return a tuple of (requests.Response, access_token)
// where you can inspect the http response, and get the updated access_token
// if it was updated (otherwise the original access_token will be returned).
func  TriggerWorkflow(accessToken string, refreshToken string, clientId string, workflowId string, subscriberId string, userId string, targets []string, actionArgs map[string]string) map[string]string {
	var res *http.Response
	// Create the query params to be sent with the request, and encode the query params
	queryParams := url.Values{}
	queryParams.Add("subscriber_id", subscriberId)
	queryParams.Add("user_id", userId)

	triggerUrl := "https://" + serverHostname + "/ibot/workflow/" + workflowId + "?" + queryParams.Encode()

	// Create a map representing the payload to be sent with teh request.  Add action_args field if actionArgs has entries.  Convert
	// the triggerPayload map into a string and then into bytes that can be sent with the request
	
	triggerPayload := map[string]string{
		"action": "invoke",
	}
	if len(actionArgs) > 0 {
		actionArgsString, err := json.Marshal(actionArgs)
		if err != nil {
			log.Error("Error converting actionArgs to JSON: ", err)
		}
		triggerPayload["action_args"] = string(actionArgsString)
	}
	if len(targets) > 0 {
		targetsString, err := json.Marshal(targets)
		if err != nil {
			log.Error("Error converting targets to JSON: ", err)
		}
		triggerPayload["target_device_ids"] = string(targetsString)
	}
	triggerPayloadString, err := json.Marshal(triggerPayload)
	if err != nil {
		log.Error("Error converting triggerPayload to JSON: ", err)
	}
	var payload = []byte(string(triggerPayloadString))

	// Create a requst to be sent with the triggerUrl and payload bytes
	req, err := http.NewRequest("POST", triggerUrl, bytes.NewBuffer(payload))
	if err != nil {
		log.Error("Error creating HTTP request in TriggerWorkflow: ", err)
	}

	// Set the headers
	req.Header.Set("User-Agent", version)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Create the client and perform the request
	client := &http.Client{}
	client.Timeout = time.Second * 30
	res, err = client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request in TriggerWorkflow: ", err)
	}

	// If you get a 401 back, retrieve a new access token and try again
	if res.StatusCode == http.StatusUnauthorized {
		fmt.Println("Got 401, retrieving a new access token")
		accessToken = updateAccessToken(refreshToken, clientId)
		req, err := http.NewRequest("POST", triggerUrl, bytes.NewBuffer(payload))
		if err != nil {
			log.Error("Error creating HTTP request in TriggerWorkflow", err)
		}
		// Set the headers again
		req.Header.Set("User-Agent", version)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		// Create the client and perform the request again
		client := &http.Client{}
		client.Timeout = time.Second * 30
		res, err = client.Do(req)
		if err != nil {
			log.Error("Error sending HTTP request in TriggerWorkflow: ", err)
		}
	}

	// Convert the respoonse body into bytes, so that it can then be converted into a string that is readable to the client
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("Error reading HTTP response in TriggerWorkflow: ", err)
	}
	// Return a map containing the response and the access token
	response := map[string]string{
		"response":     string(bytes),
		"access_token": accessToken,
	}

	defer res.Body.Close()

	return response
}

// A convenience method for getting all the details of a device.
// This will return quite a bit of data regarding device configuration and
// state. The result, if the query was successful, should have a large JSON
// dictionary.
func FetchDevice(accessToken string, refreshToken string, clientId string, subscriberId string, userId string) map[string]string {
	// Create the query params to be sent with the request, and encode the query params
	queryParams := url.Values{}
	queryParams.Add("subscriber_id", subscriberId)

	url := "https://" + serverHostname + "/relaypro/api/v1/device/" + userId + "?" + queryParams.Encode()

	// Create the client and request
	client := &http.Client{}
	client.Timeout = time.Second * 30
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Error creating HTTP request in FetchDevice: ", err)
	}

	// Set the headers and perform the request
	req.Header.Set("User-Agent", version)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := client.Do(req)
	if err != nil {
		log.Error("Error sending HTTP request in FetchDevice: ", err)
	}

	defer res.Body.Close()

	// If you get a 401 back, retrieve the new access token and try again
	if res.StatusCode == http.StatusUnauthorized {
		fmt.Println("Got 401, retrieving a new access token")
		accessToken = updateAccessToken(refreshToken, clientId)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		res, err = client.Do(req)
		if err != nil {
			log.Error("Error sending HTTP request in FetchDevice: ", err)
		}
		defer res.Body.Close()
	}

	// Convert the response body into types, so that it can then be converted into a string that is readable to the client
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error("Error reading HTTP response in FetchDevice: ", err)
	}

	// Return a map containing the response and the access token
	response := map[string]string{
		"response":     string(bytes),
		"access_token": accessToken,
	}
	return response
}
