package sessions

import (
	"encoding/json"
)

func (this *Session) Serialize() ([]byte, error) {
	// json
	return json.Marshal(this)
}

func (this *Session) Deserialize(byt []byte) (err error) {
	// json
	return json.Unmarshal(byt, this)
}
