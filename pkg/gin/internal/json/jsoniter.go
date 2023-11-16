// Copyright 2017 Bo-Yi Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !json && !go_json && !(sonic && avx && (linux || windows || darwin) && amd64)
// +build !json
// +build !go_json
// +build !sonic !avx !linux,!windows,!darwin !amd64

package json

import jsoniter "github.com/json-iterator/go"

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// Marshal is exported by gin/json package.
	Marshal = json.Marshal
	// Unmarshal is exported by gin/json package.
	Unmarshal = json.Unmarshal
	// MarshalIndent is exported by gin/json package.
	MarshalIndent = json.MarshalIndent
	// NewDecoder is exported by gin/json package.
	NewDecoder = json.NewDecoder
	// NewEncoder is exported by gin/json package.
	NewEncoder = json.NewEncoder
)
