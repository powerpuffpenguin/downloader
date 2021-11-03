package db

type Metadata struct {
	SumAll       []byte
	LastModified string

	SumOffset []byte
	Offset    int64
	Sum       []byte
}

func (md *Metadata) Reset() {
	md.LastModified = ``
	md.SumAll = nil
	md.Offset = 0
	md.SumOffset = nil
}
