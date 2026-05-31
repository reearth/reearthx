package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

// fromURLClient is used by FromURL to fetch remote files. It deliberately has
// no Client.Timeout because the success path streams res.Body to the caller,
// and a Client.Timeout would also abort that streaming read. Instead, the
// overall deadline is honored via the request context, while the transport
// bounds connection setup so a slow/hostile remote cannot pin a socket during
// dial, TLS handshake, or while waiting for response headers.
var fromURLClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

type File struct {
	Content         io.ReadCloser
	Name            string
	ContentType     string
	ContentEncoding string
	Path            string
	Size            int64
}

func FromMultipart(multipartReader *multipart.Reader, formName string) (*File, error) {
	if formName == "" {
		formName = "file"
	}

	for {
		p, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if p.FormName() != formName {
			if err := p.Close(); err != nil {
				return nil, err
			}
			continue
		}

		return &File{
			Content:     p,
			Name:        p.FileName(),
			Size:        0,
			ContentType: p.Header.Get("Content-Type"),
		}, nil
	}

	return nil, rerror.NewE(i18n.T("file not found"))
}

func FromURL(ctx context.Context, rawURL string) (*File, error) {
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, errors.New("failed to request")
	}

	// TODO: support gzip
	// req.Header.Set("Accept-Encoding", "gzip")

	res, err := fromURLClient.Do(req)
	if err != nil {
		return nil, rerror.ErrInternalBy(err)
	}

	if res.StatusCode > 300 {
		_ = res.Body.Close()
		return nil, rerror.ErrInternalBy(fmt.Errorf("status code is %d", res.StatusCode))
	}

	ct := res.Header.Get("Content-Type")
	ce := res.Header.Get("Content-Encoding")
	if ce != "" && ce != "gzip" {
		_ = res.Body.Close()
		return nil, fmt.Errorf("unsupported content encoding: %s", ce)
	}
	fs, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)

	fn := path.Base(URL.Path)
	_, m, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
	if err == nil && m["filename"] != "" {
		fn = m["filename"]
	}

	return &File{
		Content:         res.Body,
		Name:            fn,
		ContentType:     ct,
		ContentEncoding: ce,
		Size:            fs,
	}, nil
}
