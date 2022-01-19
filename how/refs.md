## Refs
A ref is basically a file thst stores commits
Inside the .git directory, there/s the `ref` directory. To see this, run `tree .git` in your working directory and see
the file structure. Or just run `pushd .git/refs`. Inside, there are three directories: 
- heads
- remotes
- tags

We will not be saying much about tags here, since i didn't implement it. So, we'll focus on the other two.
`heads` directory contains files that represent each of the branches that you created. 
When you run `git branch <branchname>`, _git_ creates a new file with the specified name and stores the value presently in
`.git/HEAD` inside it. `.git/HEAD` is a file that contains a symbolic reference. What this means is that it contains a relative path to a `ref`. This `ref` is that last commit made by the user. In esense, a branch always has a `parent`, and when you make a new `commit`, _git_ ass for the `ref` that is referred to currently in the `.git/HEAD` file and makes it the new _parent_ of the new commit. _Commits_ always have parents except for the first one. Enough with the theory. Hoy do we implement rferences.
A reference is implemented as a `Ref` in the `pkg` package.


```
    type Ref struct {
        path  string
        _type RefType
    }
```

The `_type` refers to the kind of `ref` this `Ref` is. It could be one of the three I listed above. Notice in the `pkg/refs.go` file that I creade a type alias `type RefType int` so I could specify the three types as constants. See below:

```
    type RefType int

    const (
        heads RefType = 0
        remotes
        tags
    )
```

There are two ways to initialize a `Ref`; one is through `InitRef`, which just takes a full path and the type of ref you're
tryint to create and returns a pointer to the new `Rf`, and the other is `RefFromSym` which is used mainly to create a `Ref` by reading the `.HEAD` file or any other file that contains a `SymRef`.

A `Ref` has two methods for now; First we can read its content with `ReadCont`. We can also update it with `UpdateCont`.
We initialize a `Ref` while creating the central `Got` object. This is useful because of the places where we need to refer to the parent of the current commit we're about to make.