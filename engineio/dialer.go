package engineio

import (
	"errors"
	"io"
	"net/url"
	"strings"

	"github.com/wudgaby/go-socket.io/engineio/packet"
	"github.com/wudgaby/go-socket.io/engineio/transport"
	"github.com/wudgaby/go-socket.io/logger"
)

// Dialer is dialer configure.
type Dialer struct {
	Transports   []transport.Transport
	ExtraHeaders map[string][]string
	Query        string
	Auth         map[string]string
}

// Dial returns a connection which dials to url with requestHeader.
func (d *Dialer) Dial(urlStr string) (Conn, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		logger.Error("parse url str:", err)

		return nil, err
	}

	query := u.Query()
	query.Set("EIO", "3")
	parseQuery, err := url.ParseQuery(d.Query)
	if err != nil {
		logger.Error("parse query err:", err)
		return nil, err
	}
	for key, vals := range parseQuery {
		query.Set(key, strings.Join(vals, ","))
	}

	u.RawQuery = query.Encode()

	var conn transport.Conn

	for i := len(d.Transports) - 1; i >= 0; i-- {
		if conn != nil {
			if closeErr := conn.Close(); closeErr != nil {
				logger.Error("close connect:", closeErr)
			}
		}

		t := d.Transports[i]

		conn, err = t.Dial(u, d.ExtraHeaders)
		if err != nil {
			logger.Error("transport dial:", err)

			continue
		}

		var params transport.ConnParameters
		if p, ok := conn.(Opener); ok {
			params, err = p.Open()
			if err != nil {
				logger.Error("open transport connect:", err)

				continue
			}
		} else {
			var pt packet.Type
			var r io.ReadCloser

			_, pt, r, err = conn.NextReader()
			if err != nil {
				continue
			}

			func() {
				defer func() {
					if closeErr := r.Close(); closeErr != nil {
						logger.Error("close connect reader:", closeErr)
					}
				}()

				if pt != packet.OPEN {
					err = errors.New("invalid open")

					return
				}

				params, err = transport.ReadConnParameters(r)
				if err != nil {
					return
				}
			}()
		}
		if err != nil {
			logger.Error("transport dialer:", err)

			continue
		}

		ret := &client{
			conn:      conn,
			params:    params,
			transport: t.Name(),
			close:     make(chan struct{}),
		}

		go ret.serve()

		return ret, nil
	}

	return nil, err
}
