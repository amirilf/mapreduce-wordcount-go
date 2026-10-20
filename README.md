# mapreduce-wordcount-go

Word count built on the MapReduce pattern, simulated on a single machine with
goroutines and channels in Go. Each stage (Map, Shuffle/Group, Reduce) runs in
independent goroutines that talk only through channels.
