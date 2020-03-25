---
date: 2018-12-09
title: Over-the-top optimisations with Nim
description: Sometimes its fun to just abandon good practice and make something *fast*.
area: optimisation
slug: over-the-top-optimisations-in-nim

resources:
  - src: advent-of-code.png
    name: Christmas Tree from Advent of Code 2005
    params:
      default: true
  - src: logo.jpg
    name: The Nim logo
---

{{< figure "right" "Christmas Tree from Advent of Code 2005" >}}

For the past few years I've been taking part in
[Eric Wastl's](https://twitter.com/ericwastl)
[Advent of Code](https://adventofcode.com/), a coding challenge that provides
a 2-part problem each day from the 1st of December through to Christmas Day.
The puzzles are always interesting — especially as they get progressively
harder — and there's an awesome community of folks that share their solutions
in a huge variety of languages.

To up the ante somewhat, [Shane](https://dataforce.org.uk/) and I usually
have a little informal competition to see who can write the most performant
code. This year, though, Shane went massively overboard and wrote an entire
[benchmarking suite and webapp](https://blog.dataforce.org.uk/2018/08/advent-of-code-benchmarking/)
to measure our performance, which I took as an invitation and personal
challenge to try to beat him every single day.

For the past three years I'd used Python exclusively, as its vast standard
library and awesome syntax lead to quick and elegant solutions. Unfortunately
it stands no chance, at least on the earlier puzzles, of beating the speed
of Shane's preferred language of PHP. For a while I consoled myself with the
notion that once the challenges get more complicated I'd be in with a shot,
but after the third or fourth time that Shane's solution finished before
the Python interpreter even started[^1] I decided I'd have to jump ship. I
started using Nim.

<!--more-->

### Introducing Nim

[Nim](https://nim-lang.org/), formerly Nimrod, is a compiled language that
takes a lot of cues from Python. It has a very nice and familiar syntax,
a reasonable standard library, and it's *fast*. I'd thought about learning
it before but didn't really have anything suitable to use it on, until now.
The code I used for my day one part one answer looks like this in Nim:

{{< highlight nim >}}
import math, sequtils, strutils

echo readFile("data/01.txt").strip.splitLines.map(parseInt).sum
{{< / highlight >}}

It's a one liner that Python would be proud of. The difference with Nim,
though, is that this compiles down to C, and from there you get all
the benefits of an optimising C compiler and linker. You end up with
a blazingly fast stand-alone binary.

### Losing my marbles

[Day 9](https://adventofcode.com/2018/day/9) of this year's Advent of
Code proved interesting to optimise, and I'm going to walk through some
of the steps I took and their impact. I'm in no way a Nim expert and
this is for a program that will be ran once and then thrown away, so
please don't take this too much to heart.

Day 9 presents a marble game played by Santa's elves, whereby marbles
with increasing values are added to a circle according to certain
rules; every 23rd marble is special and the elf playing it gets to
keep that one and also pick up a marble a certain number of places
away. The winner is the one with the highest marble value at the end.
It doesn't sound like a particularly thrilling game, but as far as
I can tell there's no way to easily predict the winner without
simulating it step-by-step so it makes for an interesting problem.

### Naive solution: over 10 minutes

My puzzle input called for a game with 72,104 marbles. My initial approach was
to use a sequence (similar to a list) to store the values of the marbles as
they're added to the circle. This got an answer for part 1 in a about 10
seconds and put me at number 124 on the global leaderboard for fastest
completion. Unfortunately, when part 2 was revealed it asked me to calculate
the result if there were 7,210,400 marbles in play.

Obviously a puzzle 100x larger would take at least 100x longer to run, and
almost certainly a lot more than that. There isn't a way to calculate the
advance stages more quickly, so the only thing to be done is to make it
run a lot faster. Seven million iterations isn't really *that* much of a
burden for a modern CPU: for the code to be running this slowly the
execution time of some of the operations must be scaling with the number
of marbles. A quick look through the documentation reveals:

{{< highlight text >}}
proc del[T](x: var seq[T]; i: Natural) {...}
deletes the item at index i by putting x[high(x)] into position i.
This is an O(1) operation.

proc delete[T](x: var seq[T]; i: Natural) {...}
deletes the item at index i by moving x[i+1..] by one position.
This is an O(n) operation.
{{< / highlight >}}

Because we have to delete a marble at an arbitrary point and maintain the
ordering of the others, I was using the `delete()` proc which has an O(n)
runtime. The other potentially costly operation is inserting a new marble;
the documentation doesn't mention the runtime but all of the nim docs have
a direct link to the source code, and we can
[see that inserting an element requires iterating over all the elements after it,](https://github.com/nim-lang/Nim/blob/72e15ff739cc73fbf6e3090756d3f9cb3d5af2fa/lib/system.nim#L1561)
so it's also O(n) in the worse case.

### DoublyLinkedLists: ~500ms

When you need performant inserts and deletes in a list, the go-to solution
is a linked list. Because nodes store references to their neighbours
(instead of being stored consecutively in an array or list), delete and
insert operations are O(1): you simply need to change a few pointers. Nim's
[lists package](https://nim-lang.org/docs/lists.html) provides a
convenient `DoublyLinkedList` that I went ahead and used.

Instead of using the old `insert` and `delete` methods I now had my own
which simply manipulate the nodes' previous and next pointers:

{{< highlight nim >}}
func insertAfter(node: DoublyLinkedNode[int], value: int) =
    var newNode = newDoublyLinkedNode(value)
    newNode.next = node.next
    newNode.prev = node
    newNode.next.prev = newNode
    newNode.prev.next = newNode

 func remove(node: DoublyLinkedNode[int]) =
    node.prev.next = node.next
    node.next.prev = node.prev
{{< / highlight >}}

This implementation brought the runtime down to a respectable 500ms,
which handily beat Shane's PHP implementation. It was still an order of
magnitude longer than any of my other solutions, though, so I wasn't
happy yet.

### Reduced imports: ~470ms

One thing I was conscious of from trying to make Python performant was how
the number of imports can pile on to startup time. I had a couple of unused
imports that were easy to shed, and I also decided to implement my own
linked list in favour of nim's `lists` module. All this involved was
defining a type and then replacing my usages of `DoublyLinkedNode[int]`
with my new `Marble`.

{{< highlight nim >}}
type
    Marble = ref object
        next, prev: Marble
        value: int
{{< / highlight >}}

These few changes didn't have a huge impact, but I was clutching at
straws and every 30ms was a small victory.

### Inlining methods and small optimisations: ~420ms

Thinking the code was about as fast as I was going to get it, I made
a final pass to see if there were any little tweaks I could make.
First off, I added the `inline` pragma to my insert and remove methods,
to hint to the C compiler that they should be inlined. I was concerned
that the overhead of calling a function seven million times would add up,
and inlining the fairly simple operation seems reasonable. It's entirely
possible the C compiler was already doing this (they're pretty clever),
but making the hint explicit in Nim is really easy so there's nothing to lose:

{{< highlight nim >}}
func insertAfter(node: Marble, value: int) {.inline.} =
    var newNode = new(Marble)
    newNode.value = value
    newNode.next = node.next
    newNode.prev = node
    newNode.next.prev = newNode
    newNode.prev.next = newNode

func remove(node: Marble) {.inline.} =
    node.prev.next = node.next
    node.next.prev = node.prev
{{< / highlight >}}

I also made some small algorithmic tweaks. These are usually the bread
and butter of optimisations but for this problem there were only a couple
I could see:

- We only care about the current player every 23rd marble, so instead
  of tracking the player each turn we can just calculate a 23 player
  jump when needed
- Instead of testing whether the current marble is divisible by 23,
  which is potentially non-trivial for large numbers, we can use a
  separate variable that just counts down from 23 and gets reset
- Instead of calculating the boundary condition (`100 * marbles`) whenever
  it's used, we can put this in a variable and calculate it once up-front.
  (The C compiler probably handled this for us anyway)

These combination of tweaks saved another 50ms, and it seemed like there
wasn't a whole lot left that could possibly change.

### Non-reference counted objects: ~180ms

While I was pondering further improvements, Shane mentioned that he managed
to make PHP's garbage collector segfault with his solution. That got me
thinking: what would happen if Nim didn't have to worry about garbage
collecting our marbles? We have a fixed amount of them and don't need to
worry about memory leaks as the program runs for half a second and then
quits. Changing the Marble type and manually allocating memory for it
— something that is virtually impossible in languages like PHP or Python —
was trivial in Nim:

{{< highlight nim >}}
type
    Marble = object
        next, prev: ptr Marble
        value: int32

proc insertAfter(node: ptr Marble, value: int) {.inline.} =
    var newNode = cast[ptr Marble](alloc0(sizeof(Marble)))
{{< / highlight >}}

Taking the garbage collector out of the equation over doubled the performance!
Still, it was my only solution that took more than 100ms and that bothered me...

### No looking back: ~120ms

Thinking about memory allocations made me take a hard look at the structure
of the `Marble` type. Each of the seven million marbles has a previous pointer
that we only use to backtrack by a fixed amount every 23rd play, which seems
wasteful. If we reduce the amount of memory we have to allocate, we'll logically
reduce the time taken allocating it.

As the game is simulated we keep track of the "current" marble, so why not
keep track of the marble eight behind that? That would allow us to turn the
doubly-linked list into a singly-linked list and save a whole bunch of memory.
This ends up being slightly complicated as initially there aren't eight marbles,
and every 23rd play we jump the current position backwards (and without
previous pointers, we can't jump the "current minus eight" pointer backwards).

To work around these issues, I added a "trailing" pointer that gradually drifts
backwards to eight behind the current pointer as moves are played. There are
22 normal moves that each advance the current pointer by two, so there's plenty
of time for this to happen.

{{< highlight nim >}}
var
    currentTrail = current
    currentTrailDrift = 0

# When a standard move occurs:
current.next.insertAfter(i)
current = current.next.next
if currentTrailDrift == 8:
    # Keep the trail eight marbles behind the current one
    currentTrail = currentTrail.next.next
else:
    # Don't move the trail so it drifts away by two marbles
    currentTrailDrift += 2
{{< / highlight >}}

This is one of those optimisations that makes the code a bit harder to follow,
but it sliced a third of the runtime off and takes us tantalisingly close to
that 100ms threshold.

### One bulk order of memory, please: ~50ms

Thinking about memory allocations, I realised we were doing seven million small
allocations over the lifetime of the program. We know upfront how many marbles
there are going to be and will need to allocate memory for them all at some
point, so why not just do it in one big bang?

Fortunately, again, Nim lets you dive from the high-level Python-like world
down to the nitty-gritty of memory management and pointers without blinking.
Now after reading the puzzle input, I allocate a big chunk of memory (for my
input with seven million marbles this equates to around 86MB of RAM) and keep
a pointer to it:

{{< highlight nim >}}
let
    hundredMarbles = marbles * 100
    memory = alloc(MarbleSize * hundredMarbles)
{{< / highlight >}}

Then when it comes to creating a "new" Marble, we simply calculate the
position in our memory block and use it as a pointer:

{{< highlight nim >}}
proc addressOf(memory: pointer, marbleNumber: int): Marble {.inline.} =
    cast[Marble](cast[uint](memory) + cast[uint](marbleNumber * MarbleSize))

proc insertAfter(node: Marble, memory: pointer, value: int): Marble {.inline.} =
    var newNode = memory.addressOf(value)
{{< / highlight >}}

Changing to this one-time allocation more than halved the runtime of the
program, placing it firmly under the 100ms target I was aiming at. It's
particularly pleasing how little effort was required for optimisations like
this, and how you can switch from high-level Python-style code to low-level
C-style pointer manipulation.

----

You can find the full code to my solution in my [aoc-2018](https://github.com/csmith/aoc-2018)
repository. If you're not taking part in [Advent of Code](https://adventofcode.com/)
I highly recommend it, and if you've not used [Nim](https://nim-lang.org/)
it's definitely worth a look.

[^1]: PHP has always been fast to start, due to its primary use in a CGI
      environment, and the last few major versions of PHP have made its
      unbelievably blazingly fast as well, while Python unfortunately
      [has issues with startup time](https://mail.python.org/pipermail/python-dev/2018-May/153296.html)
