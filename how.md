Learn Go and Git by writing a git client in Go



# Table Of Contents

Here, I write about how I wrote this client. It will be very long, so please be ready to either sit for long or pause at some point and come back later

I'd like to talk first about some important git concepts that will help you understand how git works.
I'll refer you to the [internals]() chapter of ProGit for extensive information.

Git is basically a versioned store of data, with its objects retreivable by their keys.
To ensure the keys are unique and that the objects are intact, it uses sha1 to hash the file.

sha-1 produces a 160-bit (or 20-byte) value for any input set of data you git it. Simple. That's
the point of hashing in the first place, to create a bluprint, a digest of data of iny size. the product is
always specified, definite. For sha-1, its 160 bits. But, as you possibly know already, when a byte is formatted as strings, it takes 
two chaaracters (runes in Go, but since they're between 0-f, they're still in the ascii encoding, so their width is still 8 bytes).
In the end, what you get is a 40-character string when you hash with sha-1.

Go specifics:

Two important data structures I found out that I resorted to using every now and then:

strings.Builder
&&
bytes.Buffer
&&
bufio.Scanner
&&
bufio.Writer/Reader

strings.Builder is a a cheap way to build strings by appending bytes, strings, etc to the end of the builder. In the end, you can get your 
strings all at once. Its an easy way to keep incrementing a string you'll fnally use. And it implements io.Writer. Sleek!

bytes.Buffer is just as cool. It can be read from and written to. It implements io.Writer and io.Reader, meaning I can use
it in io.copy and in every other place I either need a writer or a reader. The only thing about it that I do not like, which is actually
a feature, but which I wished that in some case, i could hac around is that it keeps an internal state of the bytes read, and as such 
when explicitly read from it or call .Bytes(), I lose access to the bytes i've read. It' s feature, its the way a reader should work, so do not mind me.

In Go, strings are basically bytes, nothing more, no promise of being valid utf-8, unlike in Rust. So, when you're converting from a
string to a byte slice, I don't think you're doing more than wrapping it really.

bufo.Scanner is  very useful when you need to iterate over an io.reader. Youcnn set your split function or use the 
predefined ones e.g. ScanLines, ScanRunes(runes are utf-8 encoded characters in go, rune's a nice name) etc. 
I'll here show you how to use both the pre-defined scanner functions and a user-defined one I used in writing this client


bufio is a cool idea. It makes io more performant by not trying to read or write every byte at once, but by populating internal buffers
and using those buffers. It also provides extension methods, useful ones at that. Never forget to fush after a write.
//How to compress using zlib
//How to write to a File
//How to use io.Writer/Reader
//How to use 

# Error Handling


I've found out that if one wnats to make the implementing of a popular protocol easier, it is a good thing to read up other people's
blog posts, presentations, and other resources that can help  understand how others think about the protocol. The formal specification
of the git protocol was to my inexperienced quite terse in some places, and I had to rely on other resources to understand. I made sure to go
back to the original specs everytime.