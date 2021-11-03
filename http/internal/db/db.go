package db

import (
	"bytes"
	"os"
)

type DB struct {
	filename string
	Metadata
	json bool
}

func New(filename string, trunc, json bool) (db *DB, e error) {
	if trunc {
		db = &DB{
			filename: filename,
			json:     json,
		}
		return
	}

	f, e := os.Open(filename)
	if e != nil {
		if os.IsNotExist(e) {
			e = nil
			db = &DB{
				filename: filename,
				json:     json,
			}
		}
		return
	}

	var md Metadata

	e = NewDecoder(f, json).Decode(&md)
	f.Close()
	if e != nil {
		return
	}
	db = &DB{
		filename: filename,
		Metadata: md,
		json:     json,
	}
	return
}
func (db *DB) Remove() (e error) {
	return os.Remove(db.filename)
}
func (db *DB) Sync() (e error) {
	f, e := os.Create(db.filename)
	if e != nil {
		return
	}
	e = NewEncoder(f, db.json).Encode(&db.Metadata)
	if e != nil {
		f.Close()
		return
	}
	e = f.Close()
	return
}
func (db *DB) CheckSumAll(sum []byte) (e error) {
	count := len(sum)
	if count == 0 {
		return
	}
	if count != len(db.SumAll) || !bytes.Equal(sum, db.SumAll) {
		db.Reset()
		db.SumAll = sum
		e = db.Sync()
	}
	return
}
