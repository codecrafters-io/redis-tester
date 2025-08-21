# Go Test Case Formatting Guide

## Package Declaration and Imports
- Package declaration on line 1
- Empty line after package declaration
- Import statements grouped with empty line between stdlib and external packages
- Empty line after imports

## Type Definitions
- Type definitions immediately after imports with no empty line
- Struct fields are not separated by empty lines
- Empty line after each type definition

## Method Grouping
- Methods on the same receiver are grouped together without empty lines between them
- Empty line between methods with different receivers or between different logical groups

## Function Bodies
- No empty lines at the start or end of function bodies
- Empty lines used sparingly within functions to separate logical blocks
- Variable declarations at the beginning of functions are not separated by empty lines

## Control Flow
- No empty lines before or after if statements, for loops, or switch statements
- Empty lines can be used to separate distinct logical operations within loops

## Comments
- Single-line comments directly above the code they describe (no empty line)
- Multi-line comments for structs and methods follow Go conventions

## Specific Patterns

### Simple Test Cases (e.g., WaitTestCase, ZaddTestCase)
```go
type TestCase struct {
    Field1 type1
    Field2 type2
}

func (t TestCase) Run(...) error {
    // Implementation without empty lines unless separating logical blocks
    return nil
}
```

### Complex Test Cases with Multiple Methods
```go
type ComplexTestCase struct {
    Fields
}

func (t *ComplexTestCase) Method1() error {
    // Implementation
}

func (t *ComplexTestCase) Method2() error {
    // Implementation
}
```

### Multi-Step Operations
- When performing multiple related operations (like in TransactionTestCase.RunAll), no empty lines between sequential method calls
- Empty line before return statement only if it improves readability

### Error Handling
- Error checks immediately follow the operation that might fail (no empty line)
- Multiple error checks in sequence don't have empty lines between them

## Key Principles
1. Minimize empty lines within functions
2. Use empty lines to separate type definitions and method groups
3. Keep related code visually grouped together
4. Follow standard Go formatting conventions (gofmt)