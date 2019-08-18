package tool

import "net/url"

func FixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}

	baseUri, err := url.Parse(base)
	if err != nil {
		return ""
	}

	uri = baseUri.ResolveReference(uri)
	return uri.String()
}
