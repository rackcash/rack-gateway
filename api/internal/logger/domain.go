package logger

type Source struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type AppInfo struct {
	Pid       int    `json:"pid"`
	GoVersion string `json:"go_version"`
}
type LogMessage struct {
	Message  string         `json:"message"`
	LogLevel string         `json:"log_level"`
	Args     map[string]any `json:",omitempty"`
	Source   Source         `json:"source"`
	AppInfo  AppInfo        `json:"app_info"`
}

const NA = "N/A"

// log level
const (
	LL_ERROR = iota
	LL_FATAL
	LL_INFO
	LL_DEBUG
)

// log stream
const (
	LS_INVOICES = iota
	LS_FATAL
	LS_NATS
	LS_WEBHOOKS
)

type Logstream uint8
type LogLevel uint8

func (l Logstream) ToString() string {
	return [...]string{"invoices", "fatal", "nats", "webhooks"}[l]
}

func (l LogLevel) ToString() string {
	return [...]string{"ERROR", "INFO", "FATAL", "DEBUG"}[l]
}
