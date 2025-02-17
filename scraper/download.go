package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"

	"goscrape/css"
	"goscrape/htmlindex"
)

// assetProcessor is a processor of a downloaded asset that can transform
// a downloaded file content before it will be stored on disk.
type assetProcessor func(URL *url.URL, data []byte) []byte

var tagsWithReferences = []string{
	htmlindex.LinkTag,
	htmlindex.ScriptTag,
	htmlindex.BodyTag,
	htmlindex.StyleTag,
}

func (s *Scraper) downloadReferences(ctx context.Context, index *htmlindex.Index) ([]ScrapeSummary, error) {
	result := []ScrapeSummary{}

	references, err := index.URLs(htmlindex.BodyTag)
	if err != nil {
	}
	s.imagesQueue = append(s.imagesQueue, references...)

	references, err = index.URLs(htmlindex.ImgTag)
	if err != nil {
	}
	s.imagesQueue = append(s.imagesQueue, references...)

	for _, tag := range tagsWithReferences {
		references, err = index.URLs(tag)
		if err != nil {

		}

		var processor assetProcessor
		if tag == htmlindex.LinkTag {
			processor = s.cssProcessor
		}
		for _, ur := range references {
			partialResult, err := s.downloadAsset(ctx, ur, processor)
			if err != nil && errors.Is(err, context.Canceled) {
				return nil, err
			}
			if partialResult != nil {
				result = append(result, *partialResult)
			}
		}
	}

	for _, image := range s.imagesQueue {
		partialResult, err := s.downloadAsset(ctx, image, s.checkImageForRecode)
		if err != nil && errors.Is(err, context.Canceled) {
			return nil, err
		}
		if partialResult != nil {
			result = append(result, *partialResult)
		}
	}

	s.imagesQueue = nil
	return result, nil
}

// downloadAsset downloads an asset if it does not exist on disk yet.
func (s *Scraper) downloadAsset(ctx context.Context, u *url.URL, processor assetProcessor) (*ScrapeSummary, error) {
	result := &ScrapeSummary{}

	u.Fragment = ""
	urlFull := u.String()

	if !s.shouldURLBeDownloaded(u, 0, true) {
		return nil, nil
	}

	filePath := s.getFilePath(u, false)
	if s.fileExists(filePath) {
		return nil, nil
	}

	data, _, contentType, size, hash, err := s.httpDownloader(ctx, u)
	if err != nil {

		return nil, fmt.Errorf("downloading asset: %w", err)
	}

	if processor != nil {
		data = processor(u, data)
	}

	if err = s.fileWriter(filePath, data); err != nil {

	}

	result = &ScrapeSummary{
		FoundUrl:    urlFull,
		ContentType: contentType,
		Size:        size,
		FileHash:    hash,
		OutputPath:  filePath,
	}

	return result, nil
}

func (s *Scraper) cssProcessor(baseURL *url.URL, data []byte) []byte {
	urls := make(map[string]string)

	processor := func(token *css.Token, data string, u *url.URL) {
		s.imagesQueue = append(s.imagesQueue, u)

		cssPath := *u
		cssPath.Path = path.Dir(cssPath.Path) + "/"
		resolved := resolveURL(&cssPath, data, s.URL.Host, false, "")
		urls[token.Value] = resolved
	}

	cssData := string(data)
	css.Process(baseURL, cssData, processor)

	if len(urls) == 0 {
		return data
	}

	for ori, filePath := range urls {
		cssData = replaceCSSUrls(ori, filePath, cssData)
	}

	return []byte(cssData)
}
