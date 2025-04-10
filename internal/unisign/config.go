package unisign

// Application-specific constants

// MagicString is the string that will be replaced with the signature
// exactly 92 characters to match base64 encoded signature with prefix
// An ed25519 signature is 64 bytes which encodes to 88 chars in base64, plus 4 chars for "us-1" prefix
const MagicString = "us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA=="

// SignaturePrefix is added to the base64 encoded signature
const SignaturePrefix = "us1-"