// Copyright © 2022 Relay Inc.

package sdk

import (
	"net/url"
	"strings"
	log "github.com/sirupsen/logrus"
)

// The scheme used for creating a URN.
var SCHEME string = "urn"

// The root used for creating a URN.
var ROOT string = "relay-resource"

// Used to specify that the URN is for a group.
var GROUP string = "group"

// Used to specify that the URN is for an ID.
var ID string = "id"

// Used to specify that the URN is for a name.
var NAME string = "name"

// Used to specify that the URN is for a device.
var DEVICE string = "device"

// Pattern used when creating an interaction URN.
var DEVICE_PATTERN string = "?device="

// Used to specify that the URN is for an interaction.
var INTERACTION string = "interaction"

// Beginning of an interaction URN that uses the name of a device.
var INTERACTION_URI_NAME string = "urn:relay-resource:name:interaction"

// Beginning of an interaction URN that uses the ID of a device.
var INTERACTION_URI_ID string = "urn:relay-resource:id:interaction"

func construct(resourceType string, idtype string, idOrName string) string {
	return SCHEME + ":" + ROOT + ":" + idtype + ":" + resourceType + ":" + idOrName 
}

// Parses out a device name from a device or interaction URN. Returns the 
// name of the device as a string.
func ParseDeviceName(uri string) string {
	uriUnescaped, err := url.PathUnescape(uri)
	log.Error("error ", err)
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

// Parses out a device ID from a device or interaction URN. Returns the ID of the
// device as a string.
func ParseDeviceId(uri string) string {
	uriUnescaped, err := url.PathUnescape(uri)
	log.Error("error ", err)
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

// Parses out a group name from a group URN. Returns the name of the group
// as a string.
func ParseGroupName(uri string) string {
	components := strings.Split(uri, ":")
	parsedGroupName, err := url.PathUnescape(components[4])
	log.Error("error ", err)
	if(components[2] == NAME && components[3] == GROUP) {
		return parsedGroupName
	}
	log.Debug("invalid group urn")
	return ""
}

// Parses out a group ID from a group URN. Returns the ID of a group as a 
// string.
func ParseGroupId(uri string) string {
	components := strings.Split(uri, ":")
	parsedGroupId, err := url.PathUnescape(components[4])
	log.Error("error ", err)
	if(components[2] == ID && components[3] == GROUP) {
		return parsedGroupId
	}
	log.Debug("invalid group urn")
	return ""
}

// Creates a URN from a group ID. Returns the constructed URN as a string.
func GroupId(id string) string {
	return construct(GROUP, ID, url.PathEscape(id))
}

// Creates a URN from a group name. Returns the constructed URN as a string.
func GroupName(name string) string {
	return construct(GROUP, NAME, url.PathEscape(name))
}

// Creates a URN for a group member. Returns the constructed URN as a string.
func GroupMember(group string, device string) string {
	return SCHEME + ":" + ROOT + ":" + NAME + ":" + GROUP + ":" + url.PathEscape(group) + DEVICE_PATTERN + url.PathEscape(SCHEME + ":" + 
			ROOT + ":" + NAME + ":" + DEVICE + ":" + device)
}

// Creates a URN from a device ID. Returns the constructed URN as a string.
func DeviceId(id string) string {
	return construct(DEVICE, ID, url.PathEscape(id))
}

// Creates a URN from a device name. Returns the constructed URN as a string.
func DeviceName(name string) string {
	return construct(DEVICE, NAME, url.PathEscape(name))
}

// Creates a URN from an interaction name. Returns the constructed URN as a string.
func InteractionName(name string) string {
        return construct(INTERACTION, NAME, url.PathEscape(name))
}

// Checks if the URN is for an interaction. Returns true if the URN is for an interaction, false otherwise.
func IsInteractionUri(uri string) bool {
	if(strings.Contains(uri, INTERACTION_URI_NAME) || strings.Contains(uri, INTERACTION_URI_ID)) {
		return true
	}
	return false
}

// Checks if the URN is a Relay URN. Returns true if the URN is a Relay URN, false otherwise.
func IsRelayUri(uri string) bool {
	if (strings.HasPrefix(uri, SCHEME + ":" + ROOT)) {
		return true
	}
	return false
}
