/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// As a matter of policy, changes to this file should be security reviewed,
// while changes to other files are less likely to need it.

func onlyAllowVerifiedRequests(
	handler http.Handler, key *ecdsa.PublicKey, now func() time.Time) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		signature, err := base64.StdEncoding.DecodeString(r.Header.Get("For-Web-Api-Gateway-Signature"))
		if err != nil {
			ErrorInvalidHeaders.ServeHTTP(w, r)
			return
		}

		type ecdsaSignature struct {
			R, S *big.Int
		}

		ecdsaSig := new(ecdsaSignature)
		if rest, err := asn1.Unmarshal(signature, ecdsaSig); err != nil || len(rest) != 0 {
			ErrorInvalidSignature.ServeHTTP(w, r)
			return
		}
		if ecdsaSig.R.Sign() <= 0 || ecdsaSig.S.Sign() <= 0 {
			ErrorInvalidSignature.ServeHTTP(w, r)
			return
		}

		timestamp, err := strconv.ParseInt(r.Header.Get("For-Web-Api-Gateway-Request-Time-Utc"), 10, 64)
		if err != nil {
			ErrorInvalidHeaders.ServeHTTP(w, r)
			return
		}
		timeError := time.Unix(timestamp, 0).Sub(now()).Minutes()
		if timeError > 1 || timeError < -1 {
			ErrorInvalidTime.ServeHTTP(w, r)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {

			ErrorIO.ServeHTTP(w, r)
			return
		}

		reqHeadersBytes, err := json.Marshal(r.Header)
		if err != nil {
			log.Println("Could not Marshal Req Headers")
		}

		function_complete_state := &functionState{
			StartTime: time.Now(),
			Input:     fmt.Sprintf("{ 'method': 'onlyAllowVerifiedRequests', 'r.headers': {%s}, 'body': '%s', 'url': '%s' }", reqHeadersBytes, body, r.URL.String()),
			Name:      "WebApiGateway.onlyAllowVerifiedRequests",
		}

		signed := make([]byte, 0)
		signed = append(signed, []byte(r.URL.String())...)
		signed = append(signed, []byte("\n")...)
		signed = append(signed, []byte(r.Header.Get("For-Web-Api-Gateway-Request-Time-Utc"))...)
		signed = append(signed, []byte("\n")...)
		signed = append(signed, body...)

		hash := sha256.Sum256(signed)

		if !ecdsa.Verify(key, hash[:], ecdsaSig.R, ecdsaSig.S) {
			ErrorNotVerified.ServeHTTP(w, r)
			return
		}

		function_complete_state.Result = fmt.Sprintf("{ 'key': '%v', 'signed': '%v'}", key, signed)
		publish_pubsub(functionComplete, *function_complete_state)

		r2 := new(http.Request)
		*r2 = *r
		r2.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		w.Header().Set("From-Web-Api-Gateway-Was-Auth-Error", "false")

		handler.ServeHTTP(w, r2)
	}
}
