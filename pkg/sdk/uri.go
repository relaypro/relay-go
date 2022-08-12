package sdk

import (
	"net/url"
	"strings"
	"fmt"
)

var SCHEME string = "urn"

var ROOT string = "relay-resource"

var GROUP string = "group"

var ID string = "id"

var NAME string = "name"

var DEVICE string = "device"

var DEVICE_PATTERN string = "?device="

var INTERACTION_URI_NAME string = "urn:relay-resource:name:interaction"

var INTERACTION_URI_ID string = "urn:relay-resource:id:interaction"

func construct(resourceType string, idtype string, idOrName string) string {
	return SCHEME + ":" + ROOT + ":" + idtype + ":" + resourceType + ":" + idOrName 
}

func ParseDeviceName(uri string) string {
	uriUnescaped, err := url.PathUnescape(uri)
	fmt.Println("error", err)
	if(!IsInteractionUri(uriUnescaped)) {
		components := strings.Split(uriUnescaped, ":")
		if(components[2] == NAME) {
			return components[4]
		}
	} else if (IsInteractionUri(uriUnescaped)) {
		components := strings.Split(uriUnescaped, ":")
		if(components[2] == NAME && components[6] == NAME) {
			return components[8]
		}
	}
	return ""
}

func ParseDeviceId(uri string) string {
	uriUnescaped, err := url.PathUnescape(uri)
	fmt.Println("error", err)
	if(!IsInteractionUri(uriUnescaped)) {
		components := strings.Split(uriUnescaped, ":")
		if(components[2] == ID) {
			return components[4]
		}
	} else if (IsInteractionUri(uriUnescaped)) {
		components := strings.Split(uriUnescaped, ":")
		if(components[2] == ID && components[6] == ID) {
			return components[8]
		}
	}
	return ""
}


func ParseGroupName(uri string) string {
	components := strings.Split(uri, ":")
	parsedGroupName, err := url.PathUnescape(components[4])
	fmt.Println("error", err)
	if(components[2] == NAME && components[3] == GROUP) {
		return parsedGroupName
	}
	fmt.Println("invalid group urn")
	return ""
}

func ParseGroupId(uri string) string {
	components := strings.Split(uri, ":")
	parsedGroupId, err := url.PathUnescape(components[4])
	fmt.Println("error", err)
	if(components[2] == ID && components[3] == GROUP) {
		return parsedGroupId
	}
	fmt.Println("invalid group urn")
	return ""
}

func GroupId(id string) string {
	return construct(GROUP, ID, url.PathEscape(id))
}

func GroupName(name string) string {
	return construct(GROUP, NAME, url.PathEscape(name))
}


func GroupMember(group string, device string) string {
	return SCHEME + ":" + ROOT + ":" + NAME + ":" + GROUP + ":" + url.PathEscape(group) + DEVICE_PATTERN + url.PathEscape(SCHEME + ":" + 
			ROOT + ":" + NAME + ":" + DEVICE + ":" + device)
}


func DeviceId(id string) string {
	return construct(DEVICE, ID, url.PathEscape(id))
}

func DeviceName(name string) string {
	return construct(DEVICE, NAME, url.PathEscape(name))
}

func IsInteractionUri(uri string) bool {
	if(strings.Contains(uri, INTERACTION_URI_NAME) || strings.Contains(uri, INTERACTION_URI_ID)) {
		return true
	}
	return false
}

func IsRelayUri(uri string) bool {
	if (strings.HasPrefix(uri, SCHEME + ":" + ROOT)) {
		return true
	}
	return false
}