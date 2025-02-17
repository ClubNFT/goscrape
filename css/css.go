package css

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/css/scanner"
)

var cssURLRe = regexp.MustCompile(`^url\(['"]?(.*?)['"]?\)$`)

type Token = scanner.Token

type urlProcessor func(token *Token, data string, url *url.URL)

// Process the CSS data and call a processor for every found URL.
func Process(url *url.URL, data string, processor urlProcessor) {
	css := scanner.New(data)

	for {
		token := css.Next()
		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			break
		}
		if token.Type != scanner.TokenURI {
			continue
		}

		match := cssURLRe.FindStringSubmatch(token.Value)
		if match == nil {
			continue
		}

		src := match[1]
		if strings.HasPrefix(strings.ToLower(src), "data:") {
			continue // skip embedded data
		}

		u, err := url.Parse(src)
		if err != nil {
			continue
		}

		processor(token, src, u)
	}
}
