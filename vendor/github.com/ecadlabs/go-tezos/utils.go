package tezos

import (
	"encoding/json"
	"fmt"
)

/*
unmarshalHeterogeneousJSONArray is a helper function used in custom JSON
unmarshallers and intended to decode array-like objects:
	[
		"...", // object ID or hash
		{
			... // ebject with ID ommitted
		}
	]
*/
func unmarshalHeterogeneousJSONArray(data []byte, v ...interface{}) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) < len(v) {
		return fmt.Errorf("JSON array is too short, expected %d, got %d", len(v), len(raw))
	}

	for i, vv := range v {
		if err := json.Unmarshal(raw[i], vv); err != nil {
			return err
		}
	}

	return nil
}
