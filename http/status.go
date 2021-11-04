package http

import "strconv"

type Status int

const (
	StatusIdle = iota
	StatusWork
	StatusDownload
	StatusExists

	StatusCompleted
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusIdle:
		return `Idle`
	case StatusWork:
		return `Work`
	case StatusDownload:
		return `Download`
	case StatusExists:
		return `Exists`
	case StatusCompleted:
		return `Completed`
	case StatusError:
		return `Error`
	}
	return `Unknow<` + strconv.Itoa(int(s)) + `>`
}

type Notifier interface {
	Notify(status Status, e error, offset, size int64)
}
