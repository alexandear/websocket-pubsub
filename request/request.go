package request

type CommandType string

const (
	CommandSubscribe   CommandType = "SUBSCRIBE"
	CommandUnsubscribe CommandType = "UNSUBSCRIBE"
)

type Command struct {
	CommandType CommandType `json:"command"`
}
