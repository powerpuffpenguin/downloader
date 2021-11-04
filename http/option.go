package http

import (
	"context"
	"hash"
	"net/http"
)

var defaultOptions = options{
	client: http.DefaultClient,
	ctx:    context.Background(),
	sync:   1024 * 1024 * 5,
}

type options struct {
	client *http.Client
	header http.Header
	ctx    context.Context

	notifier Notifier

	sum  []byte
	hash hash.Hash

	json bool
	sync int64
}

type Option interface {
	apply(*options)
}
type funcOption struct {
	f func(*options)
}

func (fdo *funcOption) apply(do *options) {
	fdo.f(do)
}
func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithClient set http client
func WithClient(client *http.Client) Option {
	return newFuncOption(func(o *options) {
		if client == nil {
			o.client = http.DefaultClient
		} else {
			o.client = client
		}
	})
}
func WithHeader(header http.Header) Option {
	return newFuncOption(func(o *options) {
		o.header = header
	})
}

func WithContext(ctx context.Context) Option {
	return newFuncOption(func(o *options) {
		if ctx == nil {
			o.ctx = context.Background()
		} else {
			o.ctx = ctx
		}
	})
}
func WithNotifier(notifier Notifier) Option {
	return newFuncOption(func(o *options) {
		o.notifier = notifier
	})
}

// if hash != nil will calculate the download file hash
//
// if hash != nil and len(sum) != 0 will check exists before download, check integrity after download
func WithHash(hash hash.Hash, sum []byte) Option {
	return newFuncOption(func(o *options) {
		o.sum = sum
		o.hash = hash
	})
}
func WithJSON(json bool) Option {
	return newFuncOption(func(o *options) {
		o.json = json
	})
}

// WithSync whenever the specified length of data is downloaded, the download status is synchronized
func WithSync(sync int64) Option {
	return newFuncOption(func(o *options) {
		o.sync = sync
	})
}
