# Go File Formatting Guide

## Package Declaration and Imports

- Package declaration on line 1
- Empty line after package declaration
- Import statements grouped with empty line between stdlib and external packages
- Empty line after imports

### Good Example
```go
package internal

import (
	"fmt"
	"io"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)
```

### Bad Example (no empty lines between groups)
```go
package internal
import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/resp/value"
	"io"
)
```

## Type Definitions

- Type definitions immediately after imports
- Struct fields are not separated by empty lines, unless there's a comment that applies to a single field
- Empty lines around each type definition

### Good Example
```go
type StringAssertion struct {
	ExpectedValue string
}

type CommandAssertion struct {
	Command string
	Args    []string
}
```

### Bad Example (no empty lines between types)
```go
type StringAssertion struct {
	ExpectedValue string
}
type CommandAssertion struct {
	Command string
	Args    []string
}
```

## Function Bodies

- No empty lines at the start or end of function bodies
- Empty lines used sparingly within functions to separate logical blocks
- Variable declarations at the beginning of functions are not separated by empty lines unless there's a logical grouping

### Good Example
```go
func testEcho(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		return err
	}

	randomWord := random.RandomWord()

	commandTestCase := test_cases.SendCommandTestCase{
		Command:   "echo",
		Args:      []string{randomWord},
		Assertion: resp_assertions.NewStringAssertion(randomWord),
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		logFriendlyError(logger, err)
		return err
	}

	client.Close()

	return nil
}
```

### Bad Example (empty lines at start/end)
```go
func testEcho(stageHarness *test_case_harness.TestCaseHarness) error {

	b := redis_executable.NewRedisExecutable(stageHarness)
	// ... rest of function ...
	
	return nil

}
```

## Control Flow

- A `return` statement should always have a blank line before it, unless it's the only statement in the function / indented block

### Good Example
```go
func (a StringAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.SIMPLE_STRING && value.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected simple string or bulk string, got %s", value.Type)
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}

	return nil
}
```

### Bad Example (no blank line before final return)
```go
func (a StringAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.SIMPLE_STRING && value.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected simple string or bulk string, got %s", value.Type)
	}

	if value.String() != a.ExpectedValue {
		return fmt.Errorf("Expected %q, got %q", a.ExpectedValue, value.String())
	}
	return nil
}
```

## Indented blocks

- Any statement that triggers an indented block (like a `if` statement, `for` loop or even a struct initialization that spans multiple lines) should have empty lines around it

### Good Example
```go
func doDecodeValue(reader *bytes.Reader) (resp_value.Value, error) {
	firstByte, err := reader.ReadByte()
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: "Expected start of a new RESP2 value",
		}
	}

	switch firstByte {
	case '+':
		return decodeSimpleString(reader)
	case '-':
		return decodeError(reader)
	case ':':
		return decodeInteger(reader)
	default:
		reader.UnreadByte()

		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("%q is not valid", string(firstByte)),
		}
	}
}
```

### Bad Example (no empty lines around blocks)
```go
func doDecodeValue(reader *bytes.Reader) (resp_value.Value, error) {
	firstByte, err := reader.ReadByte()
	if err == io.EOF {
		return resp_value.Value{}, IncompleteInputError{
			Reader:  reader,
			Message: "Expected start of a new RESP2 value",
		}
	}
	switch firstByte {
	case '+':
		return decodeSimpleString(reader)
	case '-':
		return decodeError(reader)
	default:
		reader.UnreadByte()
		return resp_value.Value{}, InvalidInputError{
			Reader:  reader,
			Message: fmt.Sprintf("%q is not valid", string(firstByte)),
		}
	}
}
```

## Comments

- Single-line comments directly above the code they describe (no empty line)
- Multi-line comments for structs and methods follow Go conventions

### Good Example
```go
// Tests 'ECHO'
func testEcho(stageHarness *test_case_harness.TestCaseHarness) error {
	// Implementation here
}

// We use multiple replicas to assert whether sent commands are replicated
replicas, err := SpawnReplicas(replicaCount, stageHarness, logger, "localhost:6379")
```

### Bad Example (empty line between comment and code)
```go
// Tests 'ECHO'

func testEcho(stageHarness *test_case_harness.TestCaseHarness) error {
	// Implementation here
}
```

## Error Handling

- Error checks immediately follow the operation that might fail (no empty line)

### Good Example
```go
client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
if err != nil {
	return err
}

b := redis_executable.NewRedisExecutable(stageHarness)
if err := b.Run(); err != nil {
	return err
}
```

### Bad Example (empty line between operation and error check)
```go
client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")

if err != nil {
	return err
}
```

## Key Principles

1. Use empty lines within functions to improve readability, but only when there's indented blocks or logical grouping.
2. Keep related code visually grouped together
3. Follow standard Go formatting conventions (gofmt)
