//
// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*This package implements common tlog functions*/
package tlog

import (
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"

	rekor_v1 "github.com/sigstore/protobuf-specs/gen/pb-go/rekor/v1"
)

// ComputeLogID generates a SHA256 hash of a DER-encoded public key, which is
// the log ID of a transparency log.
func ComputeLogID(pub crypto.PublicKey) (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(pubBytes)
	return hex.EncodeToString(digest[:]), nil
}

// GetLogID returns the hex-encoded log ID from the TransparencyLogEntry.
func GetLogID(entry *rekor_v1.TransparencyLogEntry) (string, error) {
	if entry.GetLogId() == nil {
		return "", errors.New("entry missing Log ID")
	}

	if entry.LogId.GetKeyId() == nil {
		return "", errors.New("expected Key ID")
	}

	return hex.EncodeToString(entry.LogId.GetKeyId()), nil
}
