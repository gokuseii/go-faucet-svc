package types

import (
	"encoding/json"
	"math/big"
)

type Amount struct {
	big.Int
}

func (a *Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Amount) UnmarshalJSON(b []byte) error {
	var val string
	err := json.Unmarshal(b, &val)
	if err != nil {
		panic(err)
	}

	a.SetString(val, 10)
	return nil
}
