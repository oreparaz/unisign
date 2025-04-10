// Hello world example with unisign magic comment for C programs
#include <stdio.h>

// Magic comment for unisign signature
#if defined(__APPLE__) && defined(__MACH__)
    // macOS - Mach-O format requires segment,section format
    __attribute__((section("__NOTE,__unisign")))
#else
    // Linux/other - ELF format
    __attribute__((section(".note.unisign")))
#endif
const char magic_comment[] = "us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA==";

int main() {
    printf("Hello, world!\n");
    return 0;
} 