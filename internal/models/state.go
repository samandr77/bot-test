package models

type UserMode string

const (
	ModeNone UserMode = ""
	ModeGPT  UserMode = "gpt"
	ModeSora UserMode = "sora"
	ModeNano UserMode = "nano"
)

type WaitingFor string

const (
	WaitingNone       WaitingFor = ""
	WaitingSoraPrompt WaitingFor = "sora_prompt"
	WaitingSoraImage  WaitingFor = "sora_image"
	WaitingNanoPrompt WaitingFor = "nano_prompt"
	WaitingNanoImage  WaitingFor = "nano_image"
)

type SoraSettings struct {
	Prompt   string `json:"prompt"`
	ImageURL string `json:"image_url"`
	Duration int    `json:"duration"`
	Format   string `json:"format"`
	HD       bool   `json:"hd"`
}

type NanoSettings struct {
	Prompt   string `json:"prompt"`
	ImageURL string `json:"image_url"`
}

type UserState struct {
	Mode             UserMode          `json:"mode"`
	WaitingFor       WaitingFor        `json:"waiting_for"`
	Sora             SoraSettings      `json:"sora"`
	Nano             NanoSettings      `json:"nano"`
	Temperature      float64           `json:"temperature"`
	LastMessageID    int               `json:"last_message_id"`
	LastPrompt       string            `json:"last_prompt"`
	LastResponseKeys map[string]string `json:"last_response_keys"`
}

const (
	DefaultSoraDuration = 10
	DefaultSoraFormat   = "16:9"
	DefaultTemperature  = 0.7
)

func NewUserState() *UserState {
	return &UserState{
		Mode: ModeNone,
		Sora: SoraSettings{
			Duration: DefaultSoraDuration,
			Format:   DefaultSoraFormat,
			HD:       false,
		},
		Temperature:      DefaultTemperature,
		LastResponseKeys: make(map[string]string),
	}
}
