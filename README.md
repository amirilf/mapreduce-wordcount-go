# mapreduce-wordcount-go

A word counter built on the MapReduce pattern, written in Go.

It runs on one machine. Instead of many computers, it uses many goroutines,
and the goroutines talk to each other only through channels.

## Run

Needs Go 1.22 or newer.

```bash
go run -race main.go                                       # input1.txt ... input5.txt in the current directory
go run -race main.go a.txt b.txt                           # count the words of a.txt and b.txt
go run -race main.go 'data/*.txt'                          # every .txt file in data/
go run -race main.go 'data/*.txt' 'more/d*.txt' extra.txt  # every .txt in data/ & in more/ but starting with "d..." & specific extra.txt
go run -race main.go -nreduce 8 'data/*.txt'               # 8 reduce partitions (default 4)
go run -race main.go -chunk 65536 'data/*.txt'             # map chunks of at most 64 KB (default: one per file)
```

Each argument is a file name or a [glob pattern](https://pkg.go.dev/path/filepath#Match)
(`*`, `?`, `[abc]`). Quote patterns so the program expands them, not the shell.

The result is written to `output.txt` in the current directory, one word per
line, sorted by word:

```
and: 1
dog: 2
the: 4
```

If a file or pattern matches nothing, the program prints an error, skips it and
counts the other files. If nothing can be read, it exits with an error and does
not write `output.txt`.

Run the tests (with the race detector):

```bash
go test -race ./...
```

## Structure

```
main.go                     command line: flags, file patterns, output.txt
internal/mapreduce/
    mapreduce.go            Run: the whole pipeline (map, shuffle, reduce)
    split.go                cutting a file into chunks (-chunk)
    shuffle.go              partitioning and grouping
internal/wordcount/
    wordcount.go            Map, Reduce and the word splitter for word count
testdata/                   input files and expected output for the tests
```

`internal/mapreduce` knows nothing about words. It runs any `Map` and `Reduce`
function. `internal/wordcount` gives it the word-count versions.

## How it works

Example with 5 input files and `-nreduce 4`:

```
input1.txt ─► map ─┐                            ┌─► reduce 0 ─┐
input2.txt ─► map ─┤                            ├─► reduce 1 ─┤
input3.txt ─► map ─┼─► intermediate ─► shuffle ─┤             ├─► results ─► sort ─► output.txt
input4.txt ─► map ─┤                            ├─► reduce 2 ─┤
input5.txt ─► map ─┘                            └─► reduce 3 ─┘
```

The two sides are sized independently: there is **one map goroutine per file**
and **one reduce goroutine per partition** (`-nreduce`). With 100 files and the
default `-nreduce 4`, the program runs 100 map goroutines and 4 reduce
goroutines. Goroutines are not OS threads: the Go runtime runs them on a few
threads (one per CPU core), so one goroutine per file is cheap.

1. **Map.** The main goroutine reads each file and starts one goroutine for
   it. That goroutine calls `Map` and sends every `("word", "1")` pair to the
   `intermediate` channel. With `-chunk`, the file is first cut into chunks
   and every chunk gets its own map goroutine (see
   [Chunks](#chunks-optional)).
2. **Shuffle / Group.** The main goroutine reads every pair from
   `intermediate` and puts all values of the same word together, for example
   `"hello" → ["1", "1", "1"]`. Each word goes to partition
   `hash(word) % nReduce`, so a word always lands in the same partition.
3. **Reduce.** Every partition gets its own goroutine. It calls `Reduce` for
   each word in the partition (the sum of its values) and sends the result to
   the `results` channel.
4. **Output.** The main goroutine collects all results, sorts them by word and
   writes them as `word: count`.

### Concurrency

- A `sync.WaitGroup` waits for all map goroutines; then the `intermediate`
  channel is closed, which tells the shuffle that no more pairs will come.
- A second `sync.WaitGroup` does the same for the reduce goroutines and the
  `results` channel.
- No variable is shared between goroutines. Data moves only through channels.
  Files are read and errors are collected by the main goroutine only, the
  shuffle is the only code that touches the partition maps, and each reduce
  goroutine only reads its own partition after the shuffle is finished.
- `go run -race` and `go test -race` report no data races.

## Design decisions

We assume words are case-insensitive and punctuation is not part of
a word. Two cases are not covered by that rule, so we had to choose:

**Numbers are not words.** Text like `Chapter 12` or `1895` could count `12`
and `1895` as words. We count only letters, because a number is not a word in
normal language, and this is the usual rule in word-count programs.

**Contractions are one word.** The apostrophe is punctuation, so `don't` could
be split into `don` and `t`. We keep `don't`, `i'll` and `i'd` as one word,
because splitting them creates fake words like `t`, `ll` and `d`. Apostrophes
at the start or end of a word are removed, so `'hello'` becomes `hello`.

Hyphens split words: `re-use` becomes `re` and `use`.

### Chunks (optional)

We assume one map goroutine per file, and that is the default.
The MapReduce paper does it differently, it cuts the input into pieces of a
fixed size, so one big file is mapped by many workers at the same time.
`-chunk` adds that option:

| `-chunk`      | map goroutines on the 5 dataset files |
| ------------- | ------------------------------------- |
| `0` (default) | 5, one per file                       |
| `65536`       | 36                                    |
| `4096`        | 536                                   |

A chunk is at most `-chunk` bytes and always ends at a line break, so a word
is never cut in half between two map goroutines (the paper splits text input
at line boundaries too). The only chunk that can be bigger is a single line
that is longer than the chunk size. The result is the same for every chunk
size.

Chunking does not make this program faster, even on large input. See
[Performance](#performance) for the numbers and the reason.

## Example

- `input1.txt`: `the quick brown fox jumps over the lazy dog`
- `input2.txt`: `the dog barks and the fox runs away`

`output.txt`:

```
and: 1
away: 1
barks: 1
brown: 1
dog: 2
fox: 2
jumps: 1
lazy: 1
over: 1
quick: 1
runs: 1
the: 4
```

## Performance

It's just a measurement of how `-chunk` and `-nreduce`
change the run time. Normal build (no `-race`), Intel i5-10210U with 8
threads. Every run gave the same `output.txt`.

The 5 dataset files (2 MB), median of 5 runs:

| `-chunk` \ `-nreduce`  | 1      | 2      | 4      | 8      | 16     |
| ---------------------- | ------ | ------ | ------ | ------ | ------ |
| `0` (5 map goroutines) | 194 ms | 189 ms | 188 ms | 189 ms | 188 ms |
| `4096` (536)           | 258 ms | 257 ms | 257 ms | 256 ms | 264 ms |
| `65536` (36)           | 200 ms | 200 ms | 199 ms | 202 ms | 203 ms |
| `1048576` (5)          | 193 ms | 189 ms | 190 ms | 190 ms | 189 ms |

The same files copied 50 times (104 MB), median of 3 runs:

| `-chunk` \ `-nreduce`  | 1      | 2      | 4      | 8      | 16     |
| ---------------------- | ------ | ------ | ------ | ------ | ------ |
| `0` (5 map goroutines) | 8.7 s  | 8.7 s  | 8.9 s  | 8.9 s  | 9.0 s  |
| `4096` (~27,000)       | 18.7 s | 19.0 s | 18.7 s | 18.8 s | 18.8 s |
| `65536` (~1,700)       | 17.2 s | 17.4 s | 17.3 s | 17.3 s | 17.0 s |
| `1048576` (~105)       | 13.2 s | 13.0 s | 13.0 s | 13.2 s | 13.1 s |

**`-nreduce` does not matter.** There are only about 16,000 different words,
so adding up their counts is very fast. Almost all the time is spent before
the reduce stage starts.

**More chunks are slower.** Every `("word", "1")` pair goes through one
channel to the one shuffle stage. With 5 map goroutines, 5 goroutines send to
that channel. With chunks, hundreds or thousands send at the same time, and
they wait for each other on the channel's internal lock across all CPU
cores. Running on one core (`GOMAXPROCS=1`) shows this: there, nothing
competes for the channel and all chunk sizes take about the same time (15–16
seconds).

In the real MapReduce system each map worker writes to its own disk, so more
workers do not get in each other's way. Here they all share one channel, so
the fastest setting is the default: one map goroutine per file.
