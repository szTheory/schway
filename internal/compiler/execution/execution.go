package execution

import "encoding/json"

const (
	Schema0 = "lang.execution/0"
	Schema1 = "lang.execution/1"
)

type Outcome struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Event struct {
	Schema      string `json:"schema"`
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	FunctionID  string `json:"function_id"`
	Input       string `json:"input,omitempty"`
	Output      string `json:"output,omitempty"`
	SourcePlace string `json:"source_place,omitempty"`
	TargetPlace string `json:"target_place,omitempty"`
	TypeID      string `json:"type_id,omitempty"`
}

type Execution struct {
	Schema        string   `json:"schema"`
	Outcome       Outcome  `json:"outcome"`
	Events        []Event  `json:"events"`
	LiveResources []string `json:"live_resources"`
}

func CanonicalBytes(value Execution) ([]byte, error) { return json.Marshal(value) }
