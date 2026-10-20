package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type link struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type block struct {
	Kind  string `json:"kind"`
	Text  string `json:"text"`
	Links []link `json:"links,omitempty"`
}

type page struct {
	URL             string   `json:"url"`
	File            string   `json:"file"`
	SHA256          string   `json:"sha256"`
	FetchedAt       string   `json:"fetched_at"`
	Modified        string   `json:"source_modified,omitempty"`
	Title           string   `json:"title"`
	Blocks          []block  `json:"blocks"`
	Builds          []build  `json:"builds"`
	EquipmentBuilds []build  `json:"equipment_builds,omitempty"`
	Basics          []string `json:"basics"`
}

type build struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tips        string   `json:"tips"`
	Augments    []link   `json:"augments"`
	Items       []string `json:"items"`
}

func main() {
	out := flag.String("out", "data/guide-sources", "source snapshot directory; does not write to the database")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if err := run(ctx, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out string) error {
	if err := os.MkdirAll(filepath.Join(out, "raw"), 0755); err != nil {
		return err
	}
	client := &http.Client{Timeout: 40 * time.Second}
	get := func(u string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Hexis source review (public pages only)")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s: HTTP %d", u, resp.StatusCode)
		}
		return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	}
	sitemapFile := filepath.Join(out, "sitemap.xml")
	sitemap, err := os.ReadFile(sitemapFile)
	if os.IsNotExist(err) {
		sitemap, err = get("https://lolhaidou.cn/sitemap.xml")
		if err != nil {
			return err
		}
		if err := os.WriteFile(sitemapFile, sitemap, 0644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	var index struct {
		URLs []struct {
			Loc      string `xml:"loc"`
			Modified string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(sitemap, &index); err != nil {
		return err
	}
	pages := []page{}
	for _, entry := range index.URLs {
		u, err := url.Parse(entry.Loc)
		if err != nil || u.Scheme != "https" || u.Host != "lolhaidou.cn" || !strings.HasPrefix(u.Path, "/hero/") || !strings.HasSuffix(u.Path, ".html") {
			continue
		}
		name := filepath.Base(u.Path)
		file := filepath.Join(out, "raw", name)
		body, err := os.ReadFile(file)
		if os.IsNotExist(err) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(300 * time.Millisecond):
			}
			body, err = get(entry.Loc)
			if err != nil {
				return err
			}
			if err = os.WriteFile(file, body, 0644); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		p, err := parsePage(body, entry.Loc)
		if err != nil {
			return err
		}
		info, err := os.Stat(file)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(body)
		p.File, p.SHA256, p.FetchedAt, p.Modified = "raw/"+name, hex.EncodeToString(hash[:]), info.ModTime().UTC().Format(time.RFC3339), entry.Modified
		pages = append(pages, p)
		if len(pages)%20 == 0 {
			fmt.Printf("saved %d public hero pages\n", len(pages))
		}
	}
	data, err := json.MarshalIndent(pages, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "pages.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("saved %d hero pages with hashes and extracted public blocks to %s; no database changes\n", len(pages), out)
	return nil
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return ""
	}
	for _, class := range strings.Fields(attr(n, "class")) {
		if class == "pill-win-rate" || strings.HasPrefix(class, "drain-lock") {
			return ""
		}
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(textContent(c))
		if c.Data == "br" {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func parsePage(body []byte, source string) (page, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return page{}, err
	}
	p := page{URL: source, Blocks: []block{}}
	base, err := url.Parse(source)
	if err != nil {
		return p, err
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			p.Title = strings.TrimSpace(textContent(n))
		}
		if n.Type == html.ElementNode {
			if hasClass(n, "glass-card") {
				b := build{Augments: []link{}, Items: []string{}}
				var find func(*html.Node)
				find = func(c *html.Node) {
					if c.Data == "script" || c.Data == "style" {
						return
					}
					for _, class := range strings.Fields(attr(c, "class")) {
						if strings.HasPrefix(class, "drain-lock") {
							return
						}
					}
					if hasClass(c, "desc-text") {
						b.Description = strings.TrimSpace(textContent(c))
					}
					if hasClass(c, "tips-text") {
						b.Tips = strings.TrimSpace(textContent(c))
					}
					if c.Data == "div" && strings.Contains(attr(c, "style"), "font-size:15px;font-weight:700") {
						b.Name = strings.TrimSpace(textContent(c))
					}
					if hasClass(c, "pill-item") && c.Data != "a" {
						b.Items = append(b.Items, strings.TrimSpace(textContent(c)))
					}
					if c.Data == "a" && strings.Contains(attr(c, "href"), "/augment/") {
						if u, e := url.Parse(attr(c, "href")); e == nil {
							b.Augments = append(b.Augments, link{Name: strings.TrimSpace(textContent(c)), URL: base.ResolveReference(u).String()})
						}
					}
					for child := c.FirstChild; child != nil; child = child.NextSibling {
						find(child)
					}
				}
				find(n)
				if b.Description != "" || b.Tips != "" {
					p.Builds = append(p.Builds, b)
				} else if len(b.Items) > 0 {
					p.EquipmentBuilds = append(p.EquipmentBuilds, b)
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Data == "div" && strings.Contains(attr(c, "style"), "font-size:13px;color:#e2e8f0;margin-bottom:4px") {
						p.Basics = append(p.Basics, strings.TrimSpace(textContent(c)))
					}
				}
			}
			kind := ""
			for _, class := range strings.Fields(attr(n, "class")) {
				switch class {
				case "desc-text", "tips-text", "box-special", "box-trap", "sub-label", "section-title", "pill-item":
					kind = class
				}
			}
			if kind != "" {
				b := block{Kind: kind, Text: strings.TrimSpace(textContent(n))}
				var links func(*html.Node)
				links = func(c *html.Node) {
					if c.Type == html.ElementNode && c.Data == "a" {
						if u, err := url.Parse(attr(c, "href")); err == nil && attr(c, "href") != "" {
							b.Links = append(b.Links, link{Name: strings.TrimSpace(textContent(c)), URL: base.ResolveReference(u).String()})
						}
					}
					for child := c.FirstChild; child != nil; child = child.NextSibling {
						links(child)
					}
				}
				links(n)
				p.Blocks = append(p.Blocks, b)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return p, nil
}

func hasClass(n *html.Node, name string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == name {
			return true
		}
	}
	return false
}
