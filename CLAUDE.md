# Go File Formatting Guide

## Package Declaration and Imports

- Package declaration on line 1
- Empty line after package declaration
- Import statements grouped with empty line between stdlib and external packages
- Empty line after imports

## Type Definitions

- Type definitions immediately after imports
- Struct fields are not separated by empty lines, unless there's a comment that applies to a single field
- Empty lines around each type definition

## Function Bodies

- No empty lines at the start or end of function bodies
- Empty lines used sparingly within functions to separate logical blocks
- Variable declarations at the beginning of functions are not separated by empty lines unless there's a logical grouping

## Control Flow

- A `return` statement should always have a blank line before it, unless it's the only statement in the function / indented block

## Indented blocks

- Any statement that triggers an indented block (like a `if` statement, `for` loop or even a struct initialization that spans multiple lines) should have empty lines around it

## Comments

- Single-line comments directly above the code they describe (no empty line)
- Multi-line comments for structs and methods follow Go conventions

### Error Handling

- Error checks immediately follow the operation that might fail (no empty line)

## Key Principles

1. Use empty lines within functions to improve readability, but only when there's indented blocks or logical grouping.
2. Keep related code visually grouped together
3. Follow standard Go formatting conventions (gofmt)
