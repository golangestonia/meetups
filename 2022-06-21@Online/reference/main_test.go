// ====================================================================================
// Date: June 21, 2022.
// Welcome to Estonia Golang!
// =====================================================================
//
// Ensure you have Go 1.18+
//
// Official docs:
//   https://go.dev/doc/tutorial/fuzz
//   https://go.dev/doc/fuzz/
//
// Trophies:
//   https://github.com/dvyukov/go-fuzz#trophies

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/exp/slices"
)

// Basic usage of fuzzing.

func Equals(a, b []byte) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Notice the FuzzEquals folder that was created.

func FuzzEquals(f *testing.F) {
	// arguments must be correct
	f.Add([]byte{1}, []byte{1})
	f.Fuzz(func(t *testing.T, a, b []byte) {
		Equals(a, b)
	})
}

/*
go test -fuzz FuzzEquals

go test -fuzz FuzzEquals -run ZZZ

	-fuzztime:
		the total time or number of iterations that the fuzz target will be executed before exiting, default indefinitely.
	-fuzzminimizetime:
		the time or number of iterations that the fuzz target will be executed during each minimization attempt, default 60sec. You can completely disable minimization by setting -fuzzminimizetime 0 when fuzzing.
	-parallel:
		the number of fuzzing processes running at once, default $GOMAXPROCS. Currently, setting -cpu during fuzzing has no effect.
*/

// Checking for constraints.

func Add(a, b int64) int64 {
	return a + b
}

func FuzzAdd(f *testing.F) {
	f.Add(int64(1), int64(2)) // few example inputs
	f.Fuzz(func(t *testing.T, a, b int64) {
		if Add(a, b) != Add(b, a) {
			t.Errorf("a + b != b + a")
		}
		r := Add(a, b)
		if a > 0 && b > 0 && (a > r || b > r) {
			t.Errorf("positive a and b are larger than r")
		}
	})
}

/*
	Supported types

	string, []byte
	int, int8, int16, int32/rune, int64
	uint, uint8/byte, uint16, uint32, uint64
	float32, float64
	bool
*/

func Dot(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.NaN()
	}
	var t float64
	for i := range a {
		t += a[i] * b[i]
	}
	return t
}

func Reverse[T any](vs []T) {
	for i, j := 0, len(vs)-1; i < len(vs)/2; i, j = i+1, j-1 {
		vs[i], vs[j] = vs[j], vs[i]
	}
}

func float64sFromBytes(t *testing.T, b []byte) []float64 {
	if len(b)%8 != 0 {
		t.Skip()
	}
	xs := make([]float64, len(b)/8)
	for i := range xs {
		xs[i] = math.Float64frombits(binary.LittleEndian.Uint64(b[i*8:]))
	}
	return xs
}

func FuzzDot(f *testing.F) {
	f.Fuzz(func(t *testing.T, x, y []byte) {
		a := float64sFromBytes(t, x)
		b := float64sFromBytes(t, y)

		forward := Dot(a, b)
		Reverse(a)
		Reverse(b)
		reversed := Dot(a, b)
		if !math.IsNaN(forward) {
			if forward != reversed {
				t.Errorf("forward and reverse results differ %v %v; a=%#v b=%#v", forward, reversed, a, b)
			}
		}
	})
}

// Roundtrip:
//   Encode + Decode of run-length encoding
//
// <length> <bytes> <length> <bytes> ...

func EncodeStrings(vs []string) []byte {
	rs := []byte{}
	for _, v := range vs {
		var tmp [binary.MaxVarintLen64]byte
		n := binary.PutUvarint(tmp[:], uint64(len(vs)))

		rs = append(rs, tmp[:n]...)
		rs = append(rs, []byte(v)...)
	}
	return rs
}

func DecodeStrings(data []byte) ([]string, error) {
	vs := []string{}
	for len(data) > 0 {
		l, n := binary.Uvarint(data)
		data = data[:n]
		vs = append(vs, string(data[:l]))
		data = data[l:]
	}
	return vs, nil
}

func FuzzEncodingThree(f *testing.F) {
	f.Add("", "", "")
	f.Add("alpha", "beta", "gamma")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		in := []string{a, b, c}
		encoded := EncodeStrings(in)
		decoded, err := DecodeStrings(encoded)
		if err != nil {
			t.Fatalf("decoding failed: %v", err)
		}
		if !slices.Equal(in, decoded) {
			t.Fatalf("encode <-> decode failed %v != %v", in, decoded)
		}
	})
}

// Testing endpoints

type EqualsServer struct{}

type Compare struct {
	A string `json:"a"`
	B string `json:"b"`
}

func (server *EqualsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var in Compare
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range in.A {
		if in.A[i] != in.B[i] {
			http.Error(w, "this is not acceptable", http.StatusNotAcceptable)
			return
		}
	}

	w.Write([]byte("OK"))
}

func FuzzServer(f *testing.F) {
	f.Fuzz(func(t *testing.T, a, b string) {
		body, err := json.Marshal(Compare{A: a, B: b})
		if err != nil {
			t.Skip()
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body))

		(&EqualsServer{}).ServeHTTP(res, req)

		switch {
		case a == b:
			if res.Code != http.StatusOK {
				t.Error("wrong result")
			}
		case a != b:
			if res.Code != http.StatusNotAcceptable {
				t.Error("wrong result")
			}
		}
	})
}

/*
	Other tools:

	https://github.com/thepudds/fzgen
	https://github.com/catenacyber/ngolo-fuzzing

	More reading

	https://go.dev/doc/tutorial/fuzz
	https://go.dev/doc/fuzz/
	https://blog.fuzzbuzz.io/writing-effective-go-fuzz-tests/
*/
