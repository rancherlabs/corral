package corral

type Status int

const (
	StatusNew = 0 + iota
	StatusProvisioning
	StatusError
	StatusDeleting
	StatusReady
)

var statusStringMap = map[Status]string{
	StatusNew:          "NEW",
	StatusProvisioning: "PROVISIONING",
	StatusError:        "ERROR",
	StatusDeleting:     "DELETING",
	StatusReady:        "READY",
}

func (s Status) String() string {
	return statusStringMap[s]
}
