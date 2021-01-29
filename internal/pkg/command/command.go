package command

type Type string

const (
	Subscribe      Type = "SUBSCRIBE"
	Unsubscribe    Type = "UNSUBSCRIBE"
	NumConnections Type = "NUM_CONNECTIONS"
)
