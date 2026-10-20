package mapreduce

import "hash/fnv"

// partition returns the reduce partition of key: hash(key) % nReduce.
func partition(key string, nReduce int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % uint32(nReduce))
}

// shuffle groups the values of every key received from in until it is
// closed, and splits the keys into nReduce partitions.
func shuffle(in <-chan KeyValue, nReduce int) []map[string][]string {
	partitions := make([]map[string][]string, nReduce)
	for i := range partitions {
		partitions[i] = make(map[string][]string)
	}
	for kv := range in {
		p := partitions[partition(kv.Key, nReduce)]
		p[kv.Key] = append(p[kv.Key], kv.Value)
	}
	return partitions
}
