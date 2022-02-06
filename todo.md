- use io.MultiReader to combine bytes.Reader and strings.Reader and Files,  that I want toput in one reader
- use bufio instead of bytes.Buffer in places. bytes.Buffer implements all the io interfaces apart from io.loser and Seeker
- do two protocols
- handle the cli app by myself
- if code gets concurrent, beware of maps. they must be locked with mutexes

- use this style for errors:
```
    // Package level exported error:
    /var ErrSomething = errors.New("something went wrong")

    // Normally you call it just "err",
        result, err := doSomething()
        // and use err right away.

        // But if you want to give it a longer name, use "somethingError".
        var specificError error
        result, specificError = doSpecificThing()

        // ... use specificError later.
```

- make the pkg module return streams as messages instead of printing directly to os.Stdout. Its better that way if we want to be more general.