package core

// ChannelRegistry is a map from org => channel => user.
type ChannelRegistry map[string]map[string]map[string]bool

// Register adds a user to the registry.
func (cr ChannelRegistry) Register(org, channel, user string) {
	if cr == nil {
		cr = map[string]map[string]map[string]bool{}
	}
	_, hasOrg := cr[org]
	if !hasOrg {
		cr[org] = map[string]map[string]bool{}
	}

	_, hasChannel := cr[org][channel]
	if !hasChannel {
		cr[org][channel] = map[string]bool{}
	}

	_, hasUser := cr[org][channel][user]
	if !hasUser {
		cr[org][channel][user] = true
	} else {
		cr[org][channel][user] = false
	}
}

// Unregister adds a user to the registry.
func (cr ChannelRegistry) Unregister(org, channel, user string) {
	if cr == nil {
		return
	}
	_, hasOrg := cr[org]
	if !hasOrg {
		return
	}

	_, hasChannel := cr[org][channel]
	if !hasChannel {
		return
	}

	_, hasUser := cr[org][channel][user]
	if !hasUser {
		return
	}
	cr[org][channel][user] = false
}

// Has returns if a user is registered or not.
func (cr ChannelRegistry) Has(org, channel, user string) bool {
	if cr == nil {
		return false
	}
	_, hasOrg := cr[org]
	if !hasOrg {
		return false
	}
	_, hasChannel := cr[org][channel]
	if !hasChannel {
		return false
	}
	isRegistered, hasUser := cr[org][channel][user]
	if !hasUser {
		return false
	}
	return isRegistered
}

// UsersInChannel returns the registered users for a channel.
func (cr ChannelRegistry) UsersInChannel(org, channel string) (users []string) {
	if cr == nil {
		return
	}
	_, hasOrg := cr[org]
	if !hasOrg {
		return
	}

	_, hasChannel := cr[org][channel]
	if !hasChannel {
		return
	}

	for user, isRegistered := range cr[org][channel] {
		if isRegistered {
			users = append(users, user)
		}
	}

	return
}
