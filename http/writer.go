package http

import (
	"hash"

	"github.com/powerpuffpenguin/downloader/http/internal/db"
)

type writer struct {
	db            *db.DB
	notifier      Notifier
	offset        int64
	count         int64
	Sync          bool
	hash          hash.Hash
	ContentLength int64
	sync          int64
}

func newWriter(db *db.DB, notifier Notifier,
	offset int64,
	hash hash.Hash,
	sync int64,
) *writer {
	return &writer{
		db:       db,
		notifier: notifier,
		offset:   offset,
		Sync:     true,
		hash:     hash,
		sync:     sync,
	}
}
func (w *writer) Write(p []byte) (n int, err error) {
	count := int64(len(p))
	w.offset += count
	if w.Sync {
		w.doSync(count)
		w.notifier.Notify(StatusDownload, nil, w.offset, w.ContentLength)
	} else {
		w.notifier.Notify(StatusWork, nil, w.offset, w.ContentLength)
	}
	return len(p), nil
}
func (w *writer) doSync(count int64) {
	w.count += count
	if w.count > w.sync {
		w.count = 0
		db := w.db
		db.Offset = w.offset
		db.SumOffset = w.hash.Sum(nil)
		db.Sync()
	}
}
