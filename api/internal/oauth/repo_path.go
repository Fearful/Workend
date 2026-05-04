package oauth

import (
	"net/url"
	"strings"
)

// RepoFullNameFromURL extracts the provider-side path component of a clone
// URL. Strips the protocol/host, any leading slash, any embedded credentials,
// and a trailing ".git" suffix.
//
//	https://github.com/sveltejs/kit.git           -> "sveltejs/kit"
//	https://gitlab.com/group/subgroup/proj.git    -> "group/subgroup/proj"
//	https://x-access-token:abc@github.com/o/r.git -> "o/r"
func RepoFullNameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	p := strings.TrimPrefix(u.Path, "/")
	p = strings.TrimSuffix(p, ".git")
	return p
}
