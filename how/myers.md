If the length of the string `A` is `N`, and the length of the string `B` is `M`,

### Traces:
A trace of length L is analogous to the longest common subsequence of the two strings.
An edit script made from it will have a length `D = N-L + M-L = N + M - 2L`. 

### K-diagonals
A set of lines running through the diagonals; let call them k-lines.
These lines originate from 0 where `x-y = 0`; in which case, a path that sticks to this central diagonal does not need to insert or delete anything. It is a zero-path. Extending to the left and to the right, the k-diagonal lines take the form of `x-y` at each point. This means that striking through the lowest `y` at `x=0` gives the last k-diagonal with the lowest k value; and on the right, the k-diagonal with the highest `k-value` is the line where `x` is maximum (equal to N) and `y` is zero.
What is a D-path?
A D-path is a path that has D number of insertions and deletions. i.e. the edit script contains commands with D number of symbols
By induction, a D-path is made up of a (D-1) path extended by a non-diagonal edge followed by a set of consecutive diagonals which could be zero. This single extension of the (D-1) path could go either rightways, thus increasing x by one and increasing the value of k = x-1 by 1 too; or it could go downwards, thus increasing y by one and makin k reduce by one.

Any edit sequence that touches `k-diagonal` must have at least |k| insertions and deletions
### Lemma 1
A D-path must end on diagonal k is an element of {-D, -D+2, ..., D-2, D}
Proof by induction:
A 0-path cnsists in diagonal edges, it starts and ends on diagonal 0. Suppose by induction that a D-path must end on a diagonal thats an element of {-D, -D+2, ..., D-2, D}, then a D+1 path conssts of a D-path plus one of {a horizontal path, a vertical path}, followed possibly by a snake that does not change the diagonal its in. So, this (D+1)-path would have to end either on `k+` or `k-1`. 
Which would be one of {-D +- 1, -D+2 +- 1, ..., D-2 +- 1, D +- 1}
Which is same as: {-D+1, -D-1, -D+3, ... D-3, D-1, D +1} == {-D+1, -D-1, ..., D-1, D +1};
Thus proving D+1 path still follows the rule, and as such for all natural  numbers, the rule holds.
This extends to say that if a D-path is odd, it ends on odd k-diagonals; if its even, it ends on even k-diagonal.
And the range of possible k-diagonal values it must end on is between -D and +D, in steps of 2.

A furthest-reaching D-path in a diagonal k is one of the sever D-paths that end on diagonal `k` but which has the highest row number (or colum number) of them all. It ends furthest from the origin (0,0). 

### Lemma 2
A furthest-reaching D-path on diagonal `k` is either a the furthest reaching `k-1`-path on k-diagonal `k-1` followed by an horizontal edge and a possibly zero number of diagonals; or the furthest-reaching `k+1` path followed by a vertical edge and a possibly-zero number of snakes.

Now, given the end-point of (D-1) paths, in diagonal `k+1` or `k-1` (depending on the next edge). Lt's call these endpoints (x', y') and (x'', y') respectively. `Lemma 2` gives us how to compute the endpoint of the furthest-reaching D-path in diagonal k.
This would either be:
1. (x', y'+1)
2. (x''+1, y'')
It would be one of the above two, but the furthest-reaching of them.
Also, we know from `lemma 1` that a D-path can only fall into one of `D+1` diagonals (because the number of diagonals between -D and D stepped in 2 is `D+1`). 


### Our sumple solution:
Compute all tthe endpoints of  

Each `k-diagonal` is a  line, and a `D-path` cn end on it. For each `k-diagonal`, there is a furthest-reaching `D-path`. This furthest-reaching `D-path` on the `k-diagonal` is the parent `D-path` to the next furthest-reaching `D-path`.
So, 
- we start knowing that `D-path`s can range from 0 to `N+M`, because the edit script can contain between 0 changes and `N+M`    changes.
- We know that each `D-path` has `D+1` number of `k-diagonals` it can end on, ranging from `-D` to `+D`.
- For each of these `D-paths`, starting from the least (0),  we run through the range of possible endpoints.
- For each of these endpoints `k`, i.e. `k-diagonal`, there is a furthest-reaching `D-path`, which is the mother of the next 


### The basic algorithm
```
const MAX = M+N
V := make([]intrface{}, (2 * MAX) + 1)
V[1] = 0
for D := 0; D <= MAX; D++ {
    for k := D; k <= D; K+= 2 {
        if k == -D || k != D // if it is -D, it certainly came from up. if k is not equal to D and  
    }
}
```