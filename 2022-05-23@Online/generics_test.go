// ====================================================================================
// Date: May 23, 2022.
// Welcome to Estonia Golang! Go Generics Edition.
// =====================================================================
//
// About me: Vitalii Lakusta, CTO @ Better Medicine | ex-TransferWise, ex-Starship.
// Contact: https://vlakusta.com/about
//
// At Better Medicine (bettermedicine.ai), we build AI-powered tools for radiologists
// and help them do cancer diagnostics more accurately and faster.
// We love Go and build our backend services in Go.
// =====================================================================
//
// OK...generics in Go 1.8. Let's go.
// All the code below will be shared after the meetup :)

package main

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/zyedidia/generic/stack"
	"golang.org/x/exp/constraints"
	"io"
	"io/ioutil"
	"testing"
)

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a float64, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func TestMinPreGeneric(t *testing.T) {
	fmt.Println(
		"minInt",
		minInt(1, 2),
	)

	fmt.Println(
		"minFloat",
		minFloat(1.1, 2.2),
	)
}

type Number interface {
	~int | ~float64 | ~string
}

// backup
// We define *type parameter constraint* on T, in this case int | float64
func min[T constraints.Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

func TestMinGeneric(t *testing.T) {

	// generics instantiation step
	minFloat := min[float64]
	minInt := min[int]
	minString := min[string]
	fmt.Println("minFloat", minFloat(2.2, 5.5))
	fmt.Println("minInt", minInt(2, 5))
	fmt.Println("minString", minString("a", "b"))

	fmt.Println(
		"min",
		min(2, 1),     // constraint type inference
		min(2.2, 1.1), // constraint type inference
	)

	// TODO [live code]: Inline interface{ int | float64 } -> int | float64
	// TODO [live code]: Refactor interface into named Number interface
	// TODO [live code]: Show that type parameter constraints cannot be used as the variable type (var a Number)
	// TODO [live code]: Show that type parameter constraints can contain union and methods in an interface definition
	// TODO [live code]: Show constraint type inference, remove type parameters from min usage
	// TODO [live code]: Show generics instantiation phase by instantiating min function with float64 type, use it

	// TODO [live code]: Explain the new symbol ~ in type param constraint
	type floatAlias float64
	var f floatAlias
	f = 5.5
	fmt.Println("floatAlias", min[floatAlias](2.2, f))

	// TODO [live code]: Replace type param constraint with constraints.Ordered. Walk through the constraints.go file
}

// TODO [live code]: Introduce comparable constraint
func Equal[T comparable](a, b T) bool {
	return a == b
}

// any is the alias for interface{}.
// `any` keyword can be used in any place in your go code where interface{} is used
func Print[T any](a T) {
	fmt.Println(a)
}

func TestEqual(t *testing.T) {
	fmt.Println(Equal(1, 1))
}

// TODO Pinpoint usage of keyword any instead of interface{}
// TODO Pinpoint necessity of using comparable for keys in generic map
// Values creates an array of the map values.
func Values[K comparable, V any](in map[K]V) []V {
	result := make([]V, 0, len(in))

	for _, v := range in {
		result = append(result, v)
	}

	return result
}

func TestMapGenericValues(t *testing.T) {
	aMap := make(map[string]string)
	aMap["key1"] = "value1"
	aMap["key2"] = "value2"
	fmt.Println(Values(aMap))
}

// TODO [live]: Go over math.go and map.go in samber/lo generics library (lodash-like lib).
// SumBy func example in math.go
// https://github.com/samber/lo/blob/master/math.go
// https://github.com/samber/lo/blob/master/math_test.go
// Keys func example in map.go
// https://github.com/samber/lo/blob/master/map.go
// https://github.com/samber/lo/blob/master/map_test.go

func TestMapGenericKeysFromLoLib(t *testing.T) {
	aMap := make(map[string]string)
	aMap["key1"] = "value1"
	aMap["key2"] = "value2"
	fmt.Println(lo.Keys(aMap))
}

// ================== Generics in structs ==================
type Bunch[E Number] []E

// TODO [live] Try to remove '[E]' from the method receiver to get the error
func (b Bunch[E]) Print() {
	fmt.Println(b)
}

// TODO [live code] Show that we can give arbitrary name for the type param receiver. Better stick to the same name as in struct definition though.
//  Rename back to E.
// TODO [live] demonstrate that method cannot have type parameters, but it can use the type param definition implementation from the struct
func (b Bunch[AnyName]) First() AnyName {
	// TODO [live - question to audience] - why would the below code work?
	//  Isn't AnyName (or E) a type parameter constraint and it cannot be used for variables? What's the deal here?
	//  Answer: you can think of the type here as already instantiated, it's technically not a type parameter constraint here
	var first AnyName
	first = b[0]
	return first
}

func PrintBunch[E Number](b Bunch[E]) {
	fmt.Println(b)
}

func TestBunch(t *testing.T) {
	bunch := Bunch[int]{1, 2, 3}

	PrintBunch[int](bunch) // without constraint type inference
	PrintBunch(bunch)      // with constraint type inference

	printBunchOnlyIntFunc := PrintBunch[int] // another example of generics instantiation
	printBunchOnlyIntFunc(bunch)

	bunch.Print()

	fmt.Println(bunch.First())
}

// Example of using generics in struct from github.com/zyedidia/generic library
// TODO [live] Walk through stack struct implementation in the library
func TestStack(t *testing.T) {
	st := stack.New[int]()
	st.Push(5)
	val := st.Pop()
	if val != 5 {
		t.Fatal("should be 5")
	}
}

// ================== Generics in interfaces ========================
type Processor[I any, O any] interface {
	Process(input I) (O, error)
}

// Example of an instantiation of an interface from a generic interface
type StringProcessor = Processor[string, string]

// the above is semantically equivalent to
// type StringProcessor interface {
//	Process(input string) (string ,error)
// }
type realStringProcessor struct{}

func (s realStringProcessor) Process(input string) (string, error) {
	return "processed/" + input + "/processed", nil
}

func PassProcessorResultToChannel[I any, O any](input I, resultCh chan<- O, processor Processor[I, O]) {
	out, _ := processor.Process(input)
	resultCh <- out
}

func TestProcessor(t *testing.T) {
	strProcessor := realStringProcessor{}
	expected, _ := strProcessor.Process("input")

	resultCh := make(chan string, 1)
	PassProcessorResultToChannel[string, string]("input", resultCh, strProcessor)
	if got := <-resultCh; expected != got {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

//============= GUIDELINES WHEN NOT TO USE GENERICS =======================================
// source: [GopherCon 2021: Robert Griesemer & Ian Lance Taylor - Generics!](https://youtu.be/Pa_e9EeCdy8)
// 1. When just calling a method on the type argument.
// 2. When the implementation of a common method is different for each type.
// 3. Whe the operation is different for each type, even without a method (e.g. reflection, json marshalling etc.)

// good
func ReadGoodExample(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
}

// bad
func ReadBadExample[T io.Reader](r T) ([]byte, error) {
	return ioutil.ReadAll(r)
}

func TestReadGoodAndBad(t *testing.T) {
	reader1 := bytes.NewBufferString("buffer string")
	reader2 := bytes.NewBufferString("buffer string")

	result1, _ := ReadGoodExample(reader1)
	result2, _ := ReadBadExample(reader2)
	fmt.Println(string(result1))
	fmt.Println(string(result2))
}

// ============= Resources on Go Generics =======================================
// - GopherCon 2021: Robert Griesemer & Ian Lance Taylor - Generics!
//   - https://youtu.be/Pa_e9EeCdy8
// - https://go.dev/doc/tutorial/generics
// - https://rakyll.org/generics-facilititators/
// ================================================================================
// Thank you!
// ================================================================================
// Let's have Q&A now
// You can leave feedback here: https://forms.gle/bRmzKjiUZmnCRM238
