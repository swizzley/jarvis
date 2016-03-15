package core

const (
	// PriorityHigh is for actions that have to be processed / checked first.
	PriorityHigh = 500

	// PriorityNormal is for typical actions (this is the default).
	PriorityNormal = 100

	// PriorityCatchAll is for actions that should be processed / checked last.
	PriorityCatchAll = 1
)

// Action represents an action that can be handled by Jarvis for a given message pattern.
type Action struct {
	ID             string
	MessagePattern string
	Description    string
	Passive        bool
	Handler        MessageHandler
	Priority       int
}

// ActionsByPriority sorts an action slice by the priority desc.
type ActionsByPriority []Action

// Len returns the slice length.
func (a ActionsByPriority) Len() int {
	return len(a)
}

// Swap swaps two indexes.
func (a ActionsByPriority) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ActionsByPriority) Less(i, j int) bool {
	return a[i].Priority > a[j].Priority
}
