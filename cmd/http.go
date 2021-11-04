package cmd

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"hash"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	internal_http "github.com/powerpuffpenguin/downloader/cmd/internal/http"
	downloader_http "github.com/powerpuffpenguin/downloader/http"
	"github.com/powerpuffpenguin/downloader/version"
	"github.com/spf13/cobra"
)

func init() {
	var (
		names    []string
		checksum string
		sumhex   []string
		header   []string
		json     bool
	)
	exec := App + ` http`
	cmd := &cobra.Command{
		Use:   `http`,
		Short: `http download`,
		Example: fmt.Sprintf(`  %s http https://ww.google.com
  %s https://ww.google.com/1 http https://ww.google.com/2
  %s -n file1 https://ww.google.com/1 http https://ww.google.com/2`,
			exec, exec, exec,
		),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(1)
				return
			}
			notifier := &notifier{
				Outputer: internal_http.Outputer{
					Writer: os.Stdout,
				},
			}
			opts := []downloader_http.Option{
				downloader_http.WithNotifier(notifier),
				downloader_http.WithJSON(json),
			}
			m := make(http.Header)
			for _, h := range header {
				strs := strings.SplitN(h, `=`, 2)
				if len(strs) == 2 {
					m.Add(strs[0], strs[1])
				}
			}
			if len(m) != 0 {
				opts = append(opts, downloader_http.WithHeader(m))
			}
			var hash hash.Hash
			if checksum != `` {
				hash = getHash(checksum)
				if hash == nil {
					log.Fatalln(`unknow checksum: `, checksum)
				}
			}

			for i, arg := range args {
				u, e := url.Parse(arg)
				if e != nil {
					log.Fatalln(e)
				}
				var name string
				if i < len(names) {
					name = names[i]
				} else {
					name = path.Base(path.Clean(u.Path))
					if name == `` || name == `/` || name == `.` {
						name = u.Host
					}
				}

				fmt.Println(`get`, u, `to`, name)
				notifier.Reset()
				worker := downloader_http.New(u.String(), name, opts...)
				if hash != nil {
					hash.Reset()
					var sum []byte
					if i < len(sumhex) {
						sum, e = hex.DecodeString(sumhex[i])
						if e != nil {
							log.Fatalln(e)
						}
					}
					worker.Hash(hash, sum)
				}
				e = worker.Serve()
				notifier.Println()
				if e != nil {
					os.Exit(1)
				}
			}
		},
	}
	flags := cmd.Flags()

	flags.StringSliceVarP(&names, `names`,
		`n`,
		nil,
		`download saved filename`,
	)
	flags.StringSliceVarP(&header, `header`,
		`H`,
		[]string{
			fmt.Sprintf(`User-Agent=Downloader/%s (%s)`, version.Version, version.Platform),
		},
		`request header`,
	)
	flags.StringVarP(&checksum, `check`,
		`c`,
		``,
		`checksum function ['MD4','MD5','SHA1','SHA224','SHA256','SHA384','SHA512','MD5SHA1','RIPEMD160','SHA3_224','SHA3_256','SHA3_384','SHA3_512','SHA512_224','SHA512_256','BLAKE2s_256','BLAKE2b_256','BLAKE2b_384','BLAKE2b_512']`,
	)
	flags.StringSliceVarP(&sumhex, `sum`,
		`s`,
		nil,
		`hash sum hex string`,
	)
	flags.BoolVarP(&json, `json`,
		`j`,
		false,
		`use json encoding to download the status file`,
	)

	rootCmd.AddCommand(cmd)
}

func getHash(name string) hash.Hash {
	switch strings.ToUpper(name) {
	case `MD4`:
		return crypto.MD4.New()
	case `MD5`:
		return crypto.MD5.New()
	case `SHA1`:
		return crypto.SHA1.New()
	case `SHA224`:
		return crypto.SHA224.New()
	case `SHA256`:
		return crypto.SHA256.New()
	case `SHA384`:
		return crypto.SHA384.New()
	case `SHA512`:
		return crypto.SHA512.New()
	case `MD5SHA1`:
		return crypto.MD5SHA1.New()
	case `RIPEMD160`:
		return crypto.RIPEMD160.New()
	case `SHA3_224`:
		return crypto.SHA3_224.New()
	case `SHA3_256`:
		return crypto.SHA3_256.New()
	case `SHA3_384`:
		return crypto.SHA3_384.New()
	case `SHA3_512`:
		return crypto.SHA3_512.New()
	case `SHA512_224`:
		return crypto.SHA512_224.New()
	case `SHA512_256`:
		return crypto.SHA512_256.New()
	case `BLAKE2s_256`:
		return crypto.BLAKE2s_256.New()
	case `BLAKE2b_256`:
		return crypto.BLAKE2b_256.New()
	case `BLAKE2b_384`:
		return crypto.BLAKE2b_384.New()
	case `BLAKE2b_512`:
		return crypto.BLAKE2b_512.New()
	}
	return nil
}

type notifier struct {
	internal_http.Outputer
	Status downloader_http.Status

	status                   downloader_http.Status
	speedWork, speedDownload *internal_http.Statistics
	offset                   int64
}

func (n *notifier) Reset() {
	n.Status = downloader_http.StatusIdle
	n.OutLine = 0
	n.speedWork = nil
	n.speedDownload = nil
	n.offset = 0
}
func (n *notifier) Notify(status downloader_http.Status, e error, offset, size int64) {
	if n.Status == status {
		n.notify(status, e, offset, size)
	} else {
		if n.Status != downloader_http.StatusIdle {
			n.Println()
		}
		n.Status = status
		n.notify(status, e, offset, size)
	}
}
func (n *notifier) notify(status downloader_http.Status, e error, offset, size int64) {
	switch status {
	case downloader_http.StatusError:
		n.PrintLine(status, ": ", e)
	case downloader_http.StatusWork:
		n.PrintLine(status, n.strWork(offset, size), n.getSpeed(offset, size, false))
		n.status = status
	case downloader_http.StatusDownload:
		n.PrintLine(status, n.strWork(offset, size), n.getSpeed(offset, size, true))
		n.status = status
	default:
		n.PrintLine(status)
	}
}
func (n *notifier) getSpeed(offset, size int64, download bool) string {
	if download {
		if n.speedWork != nil {
			n.speedWork = nil
			n.offset = 0
		}
	} else {
		if n.speedDownload != nil {
			n.speedDownload = nil
			n.offset = 0
		}
	}
	if n.offset == 0 {
		n.offset = offset
		return ""
	}
	var speed *internal_http.Statistics
	if download {
		speed = n.speedDownload
		if speed == nil {
			speed = internal_http.NewStatistics(5 * time.Second)
			n.speedDownload = speed
		}
	} else {
		speed = n.speedWork
		if speed == nil {
			speed = internal_http.NewStatistics(5 * time.Second)
			n.speedWork = speed
		}
	}

	if offset > n.offset {
		speed.Push(offset - n.offset)
		n.offset = offset
	}
	v := speed.Speed()
	if v == 0 {
		return ""
	}
	if !download {
		return fmt.Sprintf(" [%s/s]", n.strSize(v))
	}
	var eta string
	if offset < size {
		wait := float64(size - offset)
		durantion := time.Duration(wait/float64(v)*1000) * time.Millisecond
		eta = fmt.Sprintf(" %s ETA", durantion.String())
	}
	return fmt.Sprintf(" [%s/s]%s", n.strSize(v), eta)
}
func (n *notifier) strWork(offset, size int64) string {
	if offset > 0 {
		if size > 0 {
			return ": <" + n.strSize(offset) + "/" + n.strSize(size) + ">"
		} else {
			return ": <" + n.strSize(offset) + ">"
		}
	} else if size > 0 {
		return ": <0b/" + n.strSize(size) + ">"
	}
	return ""
}
func (n *notifier) strSize(size int64) string {
	if size == 0 {
		return `0b`
	}
	items := []string{"b", "kb", "m", "g"}
	strs := make([]string, 0, len(items)+1)
	var str string
	for _, sep := range items {
		if size == 0 {
			break
		}
		size, str = n.format(size, sep)
		if str != "" {
			strs = append(strs, str)
		}
	}

	return n.Join(strs, "")
}
func (n *notifier) format(size int64, sep string) (int64, string) {
	v := size % 1024
	if v == 0 {
		return size / 1024, ""
	}
	return size / 1024, fmt.Sprint(v, sep)
}
func (*notifier) Join(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[len(elems)-1])
	for i := len(elems) - 2; i > -1; i-- {
		b.WriteString(sep)
		b.WriteString(elems[i])
	}
	return b.String()
}
