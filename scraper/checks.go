package scraper

import (
	"net/url"
)

// shouldURLBeDownloaded checks whether a page should be downloaded.
// nolint: cyclop
func (s *Scraper) shouldURLBeDownloaded(url *url.URL, currentDepth uint, isAsset bool) bool {
	if url.Scheme != "http" && url.Scheme != "https" {
		return false
	}

	p := url.String()
	if url.Host == s.URL.Host {
		p = url.Path
	}
	if p == "" {
		p = "/"
	}

	if _, ok := s.processed[p]; ok { // was already downloaded or checked?
		if url.Fragment != "" {
			return false
		}
		return false
	}

	s.processed[p] = struct{}{}

	if !isAsset {
		if url.Host != s.URL.Host {
			return false
		}

		if s.config.MaxDepth != 0 && currentDepth == s.config.MaxDepth {
			return false
		}
	}

	if s.includes != nil && !s.isURLIncluded(url) {
		return false
	}
	if s.excludes != nil && s.isURLExcluded(url) {
		return false
	}

	return true
}

func (s *Scraper) isURLIncluded(url *url.URL) bool {
	for _, re := range s.includes {
		if re.MatchString(url.Path) {
			return true
		}
	}
	return false
}

func (s *Scraper) isURLExcluded(url *url.URL) bool {
	for _, re := range s.excludes {
		if re.MatchString(url.Path) {
			return true
		}
	}
	return false
}
