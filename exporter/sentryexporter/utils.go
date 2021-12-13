// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sentryexporter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

// unixNanoToTime converts UNIX Epoch time in nanoseconds
// to a Time struct.
func unixNanoToTime(u pdata.Timestamp) time.Time {
	return time.Unix(0, int64(u)).UTC()
}

func uuid() string {
	id := make([]byte, 16)
	// Prefer rand.Read over rand.Reader, see https://go-review.googlesource.com/c/go/+/272326/.
	_, _ = rand.Read(id)
	id[6] &= 0x0F // clear version
	id[6] |= 0x40 // set version to 4 (random uuid)
	id[8] &= 0x3F // clear variant
	id[8] |= 0x80 // set to IETF variant
	return hex.EncodeToString(id)
}

var (
	_debug         bool
	_debugLoadOnce sync.Once
)

func isDebug() bool {
	_debugLoadOnce.Do(func() {
		switch os.Getenv("DEBUG") {
		case "1", "true", "TRUE", "OK", "T":
			_debug = true
		}
	})

	return _debug
}

func debugf(format string, args ...interface{}) {
	if !isDebug() {
		return
	}

	fmt.Printf(format, args...)
}

type KeyGenerator interface {
	generate(key string) string
}

type fnKeyGenerator func(key string) string

func (f fnKeyGenerator) generate(key string) string {
	return f(key)
}

func intactKeyGenerator() fnKeyGenerator {
	return func(key string) string {
		return key
	}
}

func withPrefixKeyGenerator(prefix string) fnKeyGenerator {
	return func(key string) string {
		return prefix + "." + key
	}
}
