package chain

import "bytes"

type BytesSlice [][]byte

func (x BytesSlice) Len() int           { return len(x) }
func (x BytesSlice) Less(i, j int) bool { return bytes.Compare(x[i], x[j]) < 0 }
func (x BytesSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
