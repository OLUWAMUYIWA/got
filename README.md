# got
a minimal git client in go

In implementing this client, I used only the Go standard library.
I'm a Go developer interested in learning how things work at the lower level.
I'm especially interested in network protocols and distributed systems. 
If you're looking at this anywhere in the world, please check the code out and see if 
there are design or implementatioon changes I could make to improve it. I would love to learn from you.
I did well to read *Effective Go* several times over to check the quality of the code, 
but surely there are places where I could make it better. 
This client is not for production use, as it would fail you in so many ways. You may use it for any other purpose.
Hack on it, submit PRs, hopefully we can have fun together.
For devs who wish to learn from this implementation, I made the code quite readable, and punctuated it 
heavily with comments. It shouldn't be hard to find your way around it. If you need help, please let me know by raising an issue.
This [tutorial] () explains how  I implemented it. I made it a series because I wanted to go into great detail.

Special thanks to the following people, without whom I would've been unable to make this happen:

- [The Pro-Git book authors]() *For a good understanding of how git works internally, nothing is better than Pro-Git's internals chapter*
- [Ben Hoyt]() Before implementing this client, the first thing I did after reading the `internals` chapter of the book was to read through Bens Python solution. *Whenever I got stuck (and that was more than a few times), I looked again at how Ben did it in Python. Thanks a great deal!*
- [mincong](https://mincong.io/2018/04/28/git-index/) *mingcong explained the index file in a superb way. Thanks!*


## TODO
- Allow Branching
- Allow adding of deeper-nested files
- Write unit tests for all plumbers
- make asynchronous