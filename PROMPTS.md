
- Rewrite README.md in a way that a native speaker would write. Prioritize accuracy, succinctness and precise technical writing.


in file sshkey.go, add functions to read a ssh private key. the file must be ed25519 format, and the function should return a signer


in file sign.go, add functions to sign a binary buffer with a private key from ssh.Signer . The function should prepend the message to be signed with a binary struct containing:
- a fixed uint64 value
- an uint64 with the length of the message
- an uint64 that is passed as an argument to the function called `offset`


in a file replace.go, add the following functions:
- a function to find the offset of the magic string in a buffer passed as argument
- a function that replaces the magic string with another one passed as argument at an offset provided as argument to a buffer also passed as argument


in a file unisign.go, add the following main function:
- reads a ssh private key whose filename is passed as argument with the flag -k
- reads a file that is passed as second argument
- finds the offset at which the magic string us1-Cyr8XBAHq66KLsirEF3V7vs1phAcPakAt2hVgTSmUZHp happens in the passed file
- signs the file, by using the function in sign.go and the offset computed in the prior step
- replaces the magic string with the base58 encoded signature, prepended by "us-1"
- writes the resulting binary buffer in another file, with a filename ending in ".signed"