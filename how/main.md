### main

Return exit code from `cmd` because we care about the exit code of the program. It also makes the entire program easily testable.

- Because: You can’t test a function that calls os.Exit. Why? Because calling os.Exit during a test exits the test executable.

- Only call os.Exit in exactly one place, as near to the “exterior” of your application 

- The `run`/`pkg` package contains the meat of the logic of your binary. You should write this package as if it were a standalone library. It should be far removed from any thoughts of CLI, flags, etc. It should take in structured data and return errors. Pretend it might get called by some other library, or a web service, or someone else’s binary. Make as few assumptions as possible about how it’ll be used, just as you would a generic library.

[npf's advice](https://npf.io/2016/10/reusable-commands/)