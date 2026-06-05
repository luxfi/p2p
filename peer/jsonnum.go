// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package peer

import "strconv"

// Uint32 is a uint32 that JSON marshals as a quoted decimal string. Used at
// the JSON-RPC boundary where JS clients lose precision on integers > 2^53.
type Uint32 uint32

// MarshalJSON implements json.Marshaler.
func (u Uint32) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatUint(uint64(u), 10) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *Uint32) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}
	if n := len(s); n >= 2 && s[0] == '"' && s[n-1] == '"' {
		s = s[1 : n-1]
	}
	v, err := strconv.ParseUint(s, 10, 32)
	*u = Uint32(v)
	return err
}
