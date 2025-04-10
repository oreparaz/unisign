# Placeholder Package for Unisign

This package provides utilities to include the unisign magic placeholder string in your compiled binaries, making them ready for signing with the unisign tool.

## Purpose

The main purpose of this package is to ensure that the magic placeholder string required by unisign is included in your compiled binary and not optimized away by the compiler, even when using aggressive optimization flags like `-ldflags="-s -w"` or high optimization levels.

## Usage

There are several ways to use this package:

### 1. Direct function call

The simplest way is to call the `IncludePlaceholderSignatureInBinary` function from your application's main package:

```go
package main

import (
    "unisign/pkg/placeholder"
)

func main() {
    placeholder.IncludePlaceholderSignatureInBinary()
    
    // Your application logic here
}
```

### 2. Package import only

Simply importing the package is enough, as it includes an init function that ensures the magic string is included:

```go
package main

import (
    _ "unisign/pkg/placeholder" // Import for side effects only
)

func main() {
    // Your application logic here
}
```

### 3. Using the string functions

You can also use the provided functions to directly access or use the magic string:

```go
package main

import (
    "fmt"
    "unisign/pkg/placeholder"
)

func main() {
    // Get the expected length of the placeholder
    length := placeholder.GetMagicStringLength()
    fmt.Printf("Magic string length: %d\n", length)
    
    // Your application logic here
}
```

## How It Works

The package uses several techniques to prevent compiler optimizations from eliminating the "unused" string:

1. Atomic operations and unsafe pointers
2. Runtime dependencies and finalizers
3. Potential side effects through fmt package
4. Compile-time dependencies using anonymous functions
5. Package initialization

These techniques ensure that even the most aggressive compiler optimizations won't remove the magic string from your binary, making it ready for signing with unisign.

## Signing Process

After compiling your application with this package included, you can use the unisign tool to sign your binary:

```bash
unisign sign -k your_private_key.pem your_application
```

The unisign tool will locate the magic placeholder in your binary and replace it with the actual signature. 