package scraper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cornelk/gotokit/log"
)

func (s *Scraper) downloadURL(ctx context.Context, u *url.URL) ([]byte, *url.URL, string, int64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, "", 0, "", fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	if s.auth != "" {
		req.Header.Set("Authorization", s.auth)
	}

	for key, values := range s.config.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, "", 0, "", fmt.Errorf("sending HTTP request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Closing HTTP Request body failed",
				log.String("url", u.String()),
				log.Err(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, "", 0, "", fmt.Errorf("unexpected HTTP request status code %d", resp.StatusCode)
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, nil, "", 0, "", fmt.Errorf("reading HTTP request body: %w", err)
	}
	contentType := resp.Header.Get("Content-Type")

	size := int64(len(buf.Bytes()))

	hash, err := getHash(buf.Bytes())
	if err != nil {
		return nil, nil, "", 0, "", fmt.Errorf("getting hash: %w", err)
	}

	return buf.Bytes(), resp.Request.URL, contentType, size, hash, nil
}

func Headers(headers []string) http.Header {
	h := http.Header{}
	for _, header := range headers {
		sl := strings.SplitN(header, ":", 2)
		if len(sl) == 2 {
			h.Set(sl[0], sl[1])
		}
	}
	return h
}

func getHash(data []byte) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
