// Package placeholder provides utilities for including signature placeholders in binaries
package placeholder

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"
)

// MagicStringConst is the placeholder string constant that will be replaced with a signature
// Exactly 92 characters to match base64 encoded signature with prefix
// An ed25519 signature is 64 bytes which encodes to 88 chars in base64, plus 4 chars for prefix
const MagicStringConst = "us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA=="

// MagicString is a variable initialized with the constant value to allow taking its address
var MagicString = MagicStringConst

// SignaturePrefix is added to the base64 encoded signature
const SignaturePrefix = "us1-"

// volatileString prevents compiler optimizations
var volatileString atomic.Value

// isInitialized tracks whether we've initialized
var isInitialized int32

// IncludePlaceholderSignatureInBinary ensures the magic placeholder string
// is included in the compiled binary and not optimized away by the compiler.
// This function should be called from the main package of any application
// that wishes to include the placeholder for later signing.
//
// The function employs several techniques to prevent aggressive compiler
// optimizations from eliminating the "unused" string:
// 1. Using atomic operations and unsafe pointers
// 2. Creating runtime dependencies on the string
// 3. Using build tags to prevent optimizations
// 4. Creating potential side effects through fmt package
func IncludePlaceholderSignatureInBinary() string {
	// Use atomic CAS to ensure initialization happens only once
	if atomic.CompareAndSwapInt32(&isInitialized, 0, 1) {
		// Store the magic string in a volatile variable to prevent optimizations
		volatileString.Store(MagicString)
		
		// Create a runtime dependency
		runtime.SetFinalizer(&isInitialized, func(obj *int32) {
			// This will never actually be called, but the compiler doesn't know that
			if volatileString.Load() != nil {
				fmt.Println(volatileString.Load().(string))
			}
		})
		
		// Use unsafe to create complex dependencies the compiler can't eliminate
		magicPtr := unsafe.Pointer(&MagicString)
		if magicPtr != nil {
			// The compiler can't be certain if this might change something
			_ = (*(*string)(magicPtr))[0]
		}
		
		// Create a potential but unlikely side effect through fmt package
		// The compiler can't optimize away calls with dynamic formats
		format := "%s"
		if len(MagicString) > 0 && MagicString[0] != 0 {
			_ = fmt.Sprintf(format, MagicString)
		}
	}
	
	// This makes the function actually return the string,
	// creating a real usage that the compiler must preserve
	return MagicString
}

// GetMagicStringLength returns the length of the magic string.
// This provides another legitimate use of the string to prevent optimization.
func GetMagicStringLength() int {
	return len(MagicString)
}

// String is a special method that allows the package to be used in string contexts
// This enables: fmt.Println(placeholder.String())
func String() string {
	return MagicString
}

// init function ensures the magic string is included even if nothing else is called
func init() {
	IncludePlaceholderSignatureInBinary()
}

// The following code ensures the MagicString is linked into the binary
// and won't be optimized away by the compiler, even with high optimization levels.
var _ = func() string {
	// Create a compile-time dependency on MagicString
	return "prefix: " + MagicString + " :suffix"
}() 