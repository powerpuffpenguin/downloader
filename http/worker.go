package http

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/powerpuffpenguin/downloader/http/internal/db"
)

var ErrWorkerBusy = errors.New(`worker busy`)
var ErrNotMatch = errors.New(`hash not match`)

type Worker struct {
	opts     *options
	url, dst string

	err    error
	status Status
	db     *db.DB
	writer *writer
}

func New(url, dst string, opt ...Option) *Worker {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	return &Worker{
		opts:   &opts,
		url:    url,
		dst:    dst,
		status: StatusIdle,
	}
}
func (w *Worker) hash() hash.Hash {
	return md5.New()
}
func (w *Worker) notify(status Status) {
	w.status = status
	if w.opts.notifier != nil {
		w.opts.notifier.Notify(status, nil, 0, 0)
	}
}
func (w *Worker) notifyError(e error) {
	w.err = e
	w.status = StatusError
	if w.opts.notifier != nil {
		w.opts.notifier.Notify(StatusError, e, 0, 0)
	}
}

func (w *Worker) Reset(url, dst string) error {
	switch w.status {
	case StatusIdle:
		return nil
	case StatusError, StatusCompleted:
	default:
		return ErrWorkerBusy
	}

	w.url = url
	w.dst = dst
	w.err = nil
	w.db = nil
	w.writer = nil
	w.notify(StatusIdle)
	return nil
}
func (w *Worker) Hash(hash hash.Hash, sum []byte) error {
	if w.status != StatusIdle {
		return ErrWorkerBusy
	}
	w.opts.hash = hash
	w.opts.sum = sum
	return nil
}

func (w *Worker) Status() Status {
	return w.status
}
func (w *Worker) Error() error {
	return w.err
}
func (w *Worker) Serve() (e error) {
	switch w.status {
	case StatusError:
		e = w.err
		return
	case StatusIdle:
	default:
		e = ErrWorkerBusy
		return
	}
	w.notify(StatusWork)

	f, e := os.OpenFile(w.dst, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if e != nil {
		if os.IsExist(e) {
			e = w.append()
			if e == nil {
				w.notify(StatusCompleted)
			} else {
				w.notifyError(e)
			}
		}
		return
	}
	e = w.download(f)
	f.Close()
	if e == nil {
		w.notify(StatusCompleted)
	} else {
		w.notifyError(e)
	}
	return
}
func (w *Worker) responseError(resp *http.Response) error {
	body, e := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if e != nil {
		return fmt.Errorf(`%d: %v -> %e`, resp.StatusCode, resp.Status, e)
	}
	if len(body) == 0 {
		return errors.New(strconv.Itoa(resp.StatusCode) + `: ` + resp.Status)
	}
	return errors.New(strconv.Itoa(resp.StatusCode) + `: ` + resp.Status + ` -> ` + string(body))
}
func (w *Worker) download(writer io.Writer) (e error) {
	req, e := http.NewRequestWithContext(w.opts.ctx, http.MethodGet, w.url, nil)
	if e != nil {
		return
	}
	for m, k := range w.opts.header {
		req.Header[m] = k
	}
	resp, e := w.opts.client.Do(req)
	if e != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		e = w.responseError(resp)
		return
	}
	contentLength, _ := strconv.ParseInt(resp.Header.Get(`Content-Length`), 10, 64)

	m := w.db
	if m == nil {
		dir, file := filepath.Split(w.dst)
		dbname := filepath.Join(dir, `.db.d`+file)
		m, _ = db.New(dbname, true, w.opts.json)
		m.LastModified = resp.Header.Get(`Last-Modified`)
		m.SumAll = w.opts.sum
		e = m.Sync()
		if e != nil {
			return
		}
		w.db = m
	}
	db := m
	var (
		h0 = w.hash()
		h1 = w.opts.hash
		wm io.Writer
	)
	w.writer = newWriter(db, w.opts.notifier, 0, h0, w.opts.sync)
	w.writer.ContentLength = contentLength
	if h1 == nil {
		wm = io.MultiWriter(
			writer, h0,
			w.writer,
		)
	} else {
		h1.Reset()
		wm = io.MultiWriter(
			writer, h0, h1,
			w.writer,
		)
	}
	offset, e := io.Copy(wm, resp.Body)
	db.Offset = offset
	db.SumOffset = h0.Sum(nil)
	if e != nil {
		db.Sync()
		return
	}
	db.SumAll = db.SumOffset

	if h1 != nil && len(w.opts.sum) != 0 {
		v := h1.Sum(nil)
		if !bytesEqual(v, w.opts.sum) {
			db.Sync()
			e = ErrNotMatch
			return
		}
	}
	db.Remove()
	return
}
func (w *Worker) downloadTrunc() (e error) {
	f, e := os.Create(w.dst)
	if e != nil {
		return
	}
	e = w.download(f)
	f.Close()
	return
}
func (w *Worker) verifyFile(f *os.File, hash io.Writer) (e error) {
	matched, e := w.matchFile(f, hash)
	if e != nil {
		return
	} else if matched {
		w.db.Remove()
		return
	}
	if w.opts.hash != nil {
		w.opts.hash.Reset()
	}
	w.writer.Sync = true
	e = w.downloadTrunc()
	return
}
func (w *Worker) matchFile(r io.Reader, hash io.Writer) (matched bool, e error) {
	_, e = io.Copy(hash, r)
	if e != nil {
		return
	}

	if w.opts.hash != nil {
		if len(w.opts.sum) != 0 {
			sum := w.opts.hash.Sum(nil)
			if bytesEqual(sum, w.opts.sum) {
				matched = true
				return
			}
		}
	}
	return
}
func (w *Worker) append() (e error) {
	dir, file := filepath.Split(w.dst)
	dbname := filepath.Join(dir, `.db.d`+file)
	db, e := db.New(dbname, false, w.opts.json)
	if e != nil {
		return
	}
	if len(w.opts.sum) == 0 && db.Offset == 0 {
		e = w.downloadTrunc()
		return
	}
	db.CheckSumAll(w.opts.sum)
	w.db = db

	if len(w.db.SumAll) != 0 && w.opts.hash != nil && len(w.opts.sum) != 0 {
		if !bytesEqual(w.db.SumAll, w.opts.sum) {
			e = w.downloadTrunc()
			return
		}
	}

	f, e := os.OpenFile(w.dst, os.O_RDWR|os.O_APPEND, 0)
	if e != nil {
		return
	}
	defer f.Close()

	var (
		r      io.Reader = f
		h0               = w.hash()
		h1               = w.opts.hash
		writer io.Writer
	)
	w.writer = newWriter(db, w.opts.notifier, 0, h0, w.opts.sync)
	w.writer.Sync = false
	if h1 == nil {
		writer = io.MultiWriter(h0,
			w.writer,
		)
	} else {
		h1.Reset()
		writer = io.MultiWriter(h0, h1,
			w.writer,
		)
	}
	sumOffset := len(db.SumOffset) != 0 && db.Offset != 0
	if sumOffset {
		r = io.LimitReader(r, db.Offset)
	}
	offset, e := io.Copy(writer, r)
	if e != nil {
		return
	}
	if sumOffset {
		if offset != db.Offset {
			e = w.verifyFile(f, writer)
			return
		}
		sum := h0.Sum(nil)
		if !bytesEqual(sum, db.SumOffset) {
			e = w.verifyFile(f, writer)
			return
		}
	}
	matched, e := w.matchFile(f, writer)
	if e != nil {
		return
	}
	if matched {
		db.Remove()
		return
	}
	w.writer.Sync = true
	e = w.downloadRange(f, writer)
	return
}
func (w *Worker) downloadRange(f *os.File, writer io.Writer) (e error) {
	ret, e := f.Seek(0, os.SEEK_CUR)
	if e != nil {
		return
	}
	req, e := http.NewRequestWithContext(w.opts.ctx, http.MethodGet, w.url, nil)
	if e != nil {
		return
	}
	for m, k := range w.opts.header {
		req.Header[m] = k
	}
	if w.db.LastModified != `` {
		req.Header.Set(`If-Range`, w.db.LastModified)
	}
	req.Header.Set(`Range`, fmt.Sprintf(`bytes=%v-`, ret))
	resp, e := w.opts.client.Do(req)
	if e != nil {
		return
	}
	str := resp.Header.Get(`Content-Range`)
	index := strings.LastIndex(str, "/")
	if index != -1 {
		contentLength, _ := strconv.ParseInt(str[index+1:], 10, 64)
		w.writer.ContentLength = contentLength
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusPartialContent:
		e = w.appendRange(f, writer, resp)
	case http.StatusOK, http.StatusRequestedRangeNotSatisfiable:
		f.Close()
		if w.opts.hash != nil {
			w.opts.hash.Reset()
		}
		e = w.downloadTrunc()
	default:
		e = w.responseError(resp)
	}
	return
}
func (w *Worker) appendRange(f *os.File, writer io.Writer, resp *http.Response) (e error) {
	writer = io.MultiWriter(f, writer)
	_, e = io.Copy(writer, resp.Body)
	if e != nil {
		w.db.Sync()
		return
	}
	if w.opts.hash != nil && len(w.opts.sum) != 0 {
		if bytesEqual(w.opts.hash.Sum(nil), w.opts.sum) {
			w.db.Remove()
			return
		}
		f.Close()
		if w.opts.hash != nil {
			w.opts.hash.Reset()
		}
		e = w.downloadTrunc()
		return
	}
	w.db.Remove()
	return
}
func bytesEqual(a, b []byte) bool {
	return len(a) == len(b) && bytes.Equal(a, b)
}
