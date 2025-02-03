package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

func isValidDomain(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	domainRegex := regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?))*\.[a-z]{2,}$`)
	return domainRegex.MatchString(host)
}

func sanitizePath(p string) string {
	cleaned := path.Clean(p)
	trailingSlash := ""
	if strings.HasSuffix(p, "/") && !strings.HasSuffix(cleaned, "/") {
		trailingSlash = "/"
	}
	segments := strings.Split(cleaned, "/")
	for i, seg := range segments {
		if seg != "" {
			segments[i] = url.PathEscape(seg)
		}
	}
	sanitized := strings.Join(segments, "/")
	if strings.HasPrefix(p, "/") && !strings.HasPrefix(sanitized, "/") {
		sanitized = "/" + sanitized
	}
	return sanitized + trailingSlash
}

func sanitizeFragment(frag string) string {
	decoded, err := url.PathUnescape(frag)
	if err != nil {
		return frag
	}
	return url.PathEscape(decoded)
}

func ValidateURL(rawURL string) (string, error) {
	// Manually handle path encoding
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		// Handle specific parse error for invalid host characters
		if strings.Contains(err.Error(), "invalid character") {
			return "", fmt.Errorf("invalid host: %s", strings.Split(rawURL, "/")[2])
		}
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return "", errors.New("URL is missing a scheme (e.g., http, https)")
	}
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	if !allowedSchemes[parsedURL.Scheme] {
		return "", fmt.Errorf("scheme '%s' is not allowed", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return "", errors.New("URL is missing a host")
	}
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	if parsedURL.User != nil {
		return "", errors.New("URL must not contain user credentials")
	}

	if strings.ContainsAny(parsedURL.Host, " \t\r\n") {
		return "", fmt.Errorf("invalid host: %s", parsedURL.Host)
	}

	host, port, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		host = parsedURL.Host
	} else {
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return "", fmt.Errorf("invalid port: %s", port)
		}
		parsedURL.Host = net.JoinHostPort(host, port)
	}

	if !isValidDomain(host) {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	// Determine if the original rawURL contains a percent sign in the path (to decide on double escape)
	doubleEscape := false
	if strings.Index(rawURL, "%") != -1 {
		doubleEscape = true
	}

	// Re-encode path using custom sanitization logic while preserving trailing slash
	cleanedPath := path.Clean(parsedURL.Path)
	if parsedURL.Path == "" || cleanedPath == "." {
		parsedURL.Path = ""
		parsedURL.RawPath = ""
	} else {
		sanitized := sanitizePath(parsedURL.Path)
		if doubleEscape {
			// Double-escape each non-empty segment
			segments := strings.Split(sanitized, "/")
			for i, seg := range segments {
				if seg != "" {
					segments[i] = url.PathEscape(seg)
				}
			}
			sanitized = strings.Join(segments, "/")
		}
		decoded, err := url.PathUnescape(sanitized)
		if err != nil {
			return "", fmt.Errorf("failed to unescape sanitized path: %w", err)
		}
		parsedURL.RawPath = sanitized
		parsedURL.Path = decoded
	}

	parsedURL.RawQuery = parsedURL.Query().Encode()
	if parsedURL.Fragment != "" {
		parsedURL.RawFragment = url.QueryEscape(parsedURL.Fragment)
	}

	result := parsedURL.String()

	return result, nil
}