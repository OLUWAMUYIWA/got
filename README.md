

* tough call to ake, but i'm pausing this project for now. I hope i can come back to it later. I made se3veral mistakes in working on this, the most significant being that i did not write unit tests as i progressed. i also made it much bigger than i planned. now, i have other things im working on, and i find those things more interesting at the moment, and less strenous. again, i hope to return to this later. but i decided that an incomplete project is not necessarily a failed one*



# got


a minimal git client in go. 
It uses no other dependency other than `assert` for testing. Everything else is from the `go standard library`
I implement algorithms like `crc`, `myer's diff`, `x-delta`, and the `git protocol (http and ssh)` from scratch.
I took testing and error handling a little seriously for a side project, mostly because I wanted to learn how tose things work in `go`.
I also witched from `go 1.6` to `go 1.18beta` to take advantage of the new `generics` introduced in `go`. The two major things i wwish `go` had, compared to `rust` are `iterators` and `macros`, now that `go` has `generics`

In implementing this client, I used only one external package.
I'm a Go developer interested in learning how things work at the lower level.
I'm especially interested in network protocols and distributed systems. 
If you're looking at this anywhere in the world, please check the code out and see if 
there are design or implementatioon changes I could make to improve it. I would love to learn from you.
I did well to read *Effective Go* several times over to check the quality of the code, 
but surely there are places where I could make it better. 
This client is not for production use, as it would fail you in so many ways. You may use it to learn, however.
Hack on it, submit PRs, hopefully we can have fun together.
For devs who wish to learn from this implementation, I made the code quite readable, and punctuated it 
heavily with comments. It shouldn't be hard to find your way around it. If you need help, please let me know by raising an issue.
This [tutorial] () explains how  I implemented it. I made it a series because I wanted to go into great detail.

Special thanks to the following people, without whom I would've been unable to make this happen:

- [The Pro-Git book authors]() *For a good understanding of how git works internally, nothing is better than Pro-Git's internals chapter*
- [Ben Hoyt]() Before implementing this client, the first thing I did after reading the `internals` chapter of the book was to read through Bens Python solution. *Whenever I got stuck (and that was more than a few times), I looked again at how Ben did it in Python. Thanks a great deal!*
- James Coglans [blog](https://blog.jcoglan.com/2017/02/12/the-myers-diff-algorithm-part-1/) is a goldmine! I stumbled on his blog while reading an article about Rust. It turns out that he implemented a sound
git client in Ruby and posted explanations of key problems he solved. I read and internalised his detailed explanation of diffs. Thank you 
James!
-Ben Johnson is another guy whose blog I stumbled upon while writing this project. His articles are not about `git`. They're about `go`.
He made them clear, precise, and easily digestible. I found myself going back to the finished parts of the code to restructure, optimize, and use better `go` idioms. Check it out his [walkthroughs](https://www.gobeyond.dev/tag/go-walkthrough/)
- [mincong](https://mincong.io/2018/04/28/git-index/) *mingcong explained the index file in a superb way. Thanks!*
-[ChimeraCoder](https://adityamukerjee.net/) wrote a git client at the recurse center

## TODO
- Allow Branching
- Allow adding of deeper-nested files
- Write unit tests for all plumbers
- make asynchronous

