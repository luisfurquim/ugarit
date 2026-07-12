package ugarit

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"iter"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/luisfurquim/goose"
)

// Goose is the log level controller for this package.
// Log levels: 1=error warnings, 2=important general messages,
// 3=low-noise debug, 4=noisy debug, 5=very verbose debug,
// 6=debug with sensitive information.
var Goose goose.Alert

// DocMeta holds metadata for a document entry in an epub.
type DocMeta struct {
	Path     string // OPF-relative href of the document
	MimeType string
	InTOC    bool // true if the document appears in the Table of Contents
}

// BookReader provides read access to an epub file.
type BookReader interface {
	// Index returns an iterator over Table of Contents items only.
	// Key is the item title as it appears in the TOC; value is DocMeta.
	Index() iter.Seq2[string, DocMeta]

	// DocReader returns an io.Reader with the raw contents of the
	// document identified by its OPF-relative path.
	// Returns (nil, error) on failure.
	DocReader(path string) (io.Reader, error)

	// Doc parses the document at the given OPF-relative path and
	// returns a *goquery.Document and an error.
	Doc(path string) (*goquery.Document, error)

	// Docs returns an iterator over all items in the epub manifest.
	// Key is the item title (from TOC) or its manifest ID; value is DocMeta.
	Docs() iter.Seq2[string, DocMeta]
}

// --- internal XML structures for epub parsing ---

type epubContainerXML struct {
	XMLName   xml.Name              `xml:"container"`
	Rootfiles []epubContainerRootfile `xml:"rootfiles>rootfile"`
}

type epubContainerRootfile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

type epubOPFPackage struct {
	XMLName  xml.Name         `xml:"package"`
	Version  string           `xml:"version,attr"`
	Manifest []epubOPFItem    `xml:"manifest>item"`
	Spine    epubOPFSpine     `xml:"spine"`
}

type epubOPFItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type epubOPFSpine struct {
	Toc string `xml:"toc,attr"`
}

type epubNCX struct {
	XMLName xml.Name       `xml:"ncx"`
	Points  []epubNavPoint `xml:"navMap>navPoint"`
}

type epubNavPoint struct {
	Label   string `xml:"navLabel>text"`
	Content struct {
		Src string `xml:"src,attr"`
	} `xml:"content"`
	Points []epubNavPoint `xml:"navPoint"`
}

// --- reader implementation ---

type epubDocEntry struct {
	title   string  // from TOC or manifest ID
	zipPath string  // full path inside the zip archive
	meta    DocMeta // Path is OPF-relative
}

type epubReader struct {
	zr         *zip.Reader
	rootFolder string
	entries    []epubDocEntry
	byZipPath  map[string]int // zipPath -> index in entries
}

// NewReader creates a BookReader by reading and parsing the epub from r.
// Because epub files are ZIP archives (requiring random access), the entire
// content of r is buffered into memory.
func NewReader(r io.Reader) (BookReader, error) {
	Goose.Logf(3, "NewReader: reading epub content into memory\n")

	data, err := io.ReadAll(r)
	if err != nil {
		Goose.Logf(1, "NewReader: error reading input: %s\n", err)
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		Goose.Logf(1, "NewReader: error opening epub as zip: %s\n", err)
		return nil, err
	}

	// Step 1: locate the OPF rootfile via META-INF/container.xml
	rootfilePath, err := epubParseContainerXML(zr)
	if err != nil {
		return nil, err
	}
	Goose.Logf(3, "NewReader: OPF rootfile: %s\n", rootfilePath)

	// The root folder is the directory portion of the rootfile path.
	rootFolder := path.Dir(rootfilePath)
	if rootFolder == "." {
		rootFolder = ""
	}
	Goose.Logf(4, "NewReader: root folder: %q\n", rootFolder)

	// Step 2: parse the OPF package document
	pkg, err := epubParseOPF(zr, rootfilePath)
	if err != nil {
		return nil, err
	}
	Goose.Logf(2, "NewReader: epub %s — %d manifest items\n", pkg.Version, len(pkg.Manifest))

	// Build manifest ID lookup
	byID := make(map[string]epubOPFItem, len(pkg.Manifest))
	for _, item := range pkg.Manifest {
		byID[item.ID] = item
	}

	// Step 3: discover and parse the TOC to obtain title mappings.
	// tocTitles maps OPF-relative href (fragment stripped) -> display title.
	tocTitles := map[string]string{}

	// Try epub3 navigation document first (properties contains "nav").
	if navID := epubFindNavItemID(pkg.Manifest); navID != "" {
		navItem := byID[navID]
		navZipPath := epubZipPath(rootFolder, navItem.Href)
		navDir := path.Dir(navItem.Href)
		if navDir == "." {
			navDir = ""
		}
		Goose.Logf(3, "NewReader: epub3 nav document: %s\n", navZipPath)
		titles, err2 := epubParseNavXHTML(zr, navZipPath, navDir)
		if err2 != nil {
			Goose.Logf(1, "NewReader: error parsing epub3 nav: %s\n", err2)
		} else {
			tocTitles = titles
		}
	}

	// Fall back to epub2 NCX when no epub3 nav titles were found.
	if len(tocTitles) == 0 && pkg.Spine.Toc != "" {
		if ncxItem, ok := byID[pkg.Spine.Toc]; ok {
			ncxZipPath := epubZipPath(rootFolder, ncxItem.Href)
			ncxDir := path.Dir(ncxItem.Href)
			if ncxDir == "." {
				ncxDir = ""
			}
			Goose.Logf(3, "NewReader: epub2 NCX: %s\n", ncxZipPath)
			titles, err2 := epubParseNCX(zr, ncxZipPath, ncxDir)
			if err2 != nil {
				Goose.Logf(1, "NewReader: error parsing epub2 NCX: %s\n", err2)
			} else {
				tocTitles = titles
			}
		}
	}

	Goose.Logf(2, "NewReader: %d TOC entries resolved\n", len(tocTitles))

	// Step 4: build the ordered entry list from the manifest.
	entries := make([]epubDocEntry, 0, len(pkg.Manifest))
	byZipPath := make(map[string]int, len(pkg.Manifest))

	for _, mi := range pkg.Manifest {
		zp := epubZipPath(rootFolder, mi.Href)
		title, inTOC := tocTitles[mi.Href]
		if !inTOC {
			title = mi.ID
		}
		Goose.Logf(5, "NewReader: item id=%s href=%s inTOC=%v title=%q\n",
			mi.ID, mi.Href, inTOC, title)

		idx := len(entries)
		entries = append(entries, epubDocEntry{
			title:   title,
			zipPath: zp,
			meta: DocMeta{
				Path:     mi.Href,
				MimeType: mi.MediaType,
				InTOC:    inTOC,
			},
		})
		byZipPath[zp] = idx
	}

	return &epubReader{
		zr:         zr,
		rootFolder: rootFolder,
		entries:    entries,
		byZipPath:  byZipPath,
	}, nil
}

// Index returns an iterator over Table of Contents items only.
func (er *epubReader) Index() iter.Seq2[string, DocMeta] {
	return func(yield func(string, DocMeta) bool) {
		for _, e := range er.entries {
			if e.meta.InTOC {
				if !yield(e.title, e.meta) {
					return
				}
			}
		}
	}
}

// Docs returns an iterator over every item in the epub manifest.
func (er *epubReader) Docs() iter.Seq2[string, DocMeta] {
	return func(yield func(string, DocMeta) bool) {
		for _, e := range er.entries {
			if !yield(e.title, e.meta) {
				return
			}
		}
	}
}

// DocReader returns an io.Reader for the document at the OPF-relative path.
func (er *epubReader) DocReader(docPath string) (io.Reader, error) {
	zp := epubZipPath(er.rootFolder, docPath)
	Goose.Logf(4, "DocReader: %s (zip entry: %s)\n", docPath, zp)

	f := epubFindZipFile(er.zr, zp)
	if f == nil {
		Goose.Logf(1, "DocReader: not found in archive: %s\n", zp)
		return nil, fmt.Errorf("document not found in epub: %s", docPath)
	}

	rc, err := f.Open()
	if err != nil {
		Goose.Logf(1, "DocReader: error opening %s: %s\n", zp, err)
		return nil, err
	}

	Goose.Logf(3, "DocReader: opened %s\n", zp)
	return rc, nil
}

// Doc parses the document at the OPF-relative path and returns a *goquery.Document.
func (er *epubReader) Doc(docPath string) (*goquery.Document, error) {
	r, err := er.DocReader(docPath)
	if err != nil {
		return nil, err
	}
	if rc, ok := r.(io.Closer); ok {
		defer rc.Close()
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		Goose.Logf(1, "Doc: error parsing %s: %s\n", docPath, err)
		return nil, err
	}

	Goose.Logf(3, "Doc: parsed %s\n", docPath)
	return doc, nil
}

// --- internal helpers ---

func epubReadZipEntry(zr *zip.Reader, name string) ([]byte, error) {
	f := epubFindZipFile(zr, name)
	if f == nil {
		return nil, fmt.Errorf("entry not found in epub archive: %s", name)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func epubFindZipFile(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func epubParseContainerXML(zr *zip.Reader) (string, error) {
	data, err := epubReadZipEntry(zr, "META-INF/container.xml")
	if err != nil {
		return "", fmt.Errorf("invalid epub: %w", err)
	}
	var c epubContainerXML
	if err := xml.Unmarshal(data, &c); err != nil {
		return "", fmt.Errorf("invalid container.xml: %w", err)
	}
	if len(c.Rootfiles) == 0 {
		return "", errors.New("invalid epub: no rootfile declared in container.xml")
	}
	return c.Rootfiles[0].FullPath, nil
}

func epubParseOPF(zr *zip.Reader, opfPath string) (*epubOPFPackage, error) {
	data, err := epubReadZipEntry(zr, opfPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read OPF file: %w", err)
	}
	var pkg epubOPFPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("invalid OPF document: %w", err)
	}
	return &pkg, nil
}

// epubFindNavItemID returns the manifest ID of the epub3 nav document,
// identified by "nav" in its properties attribute.
func epubFindNavItemID(manifest []epubOPFItem) string {
	for _, item := range manifest {
		for _, prop := range strings.Fields(item.Properties) {
			if prop == "nav" {
				return item.ID
			}
		}
	}
	return ""
}

// epubParseNavXHTML parses an epub3 nav XHTML file and returns a map from
// OPF-relative href (without fragment) to display title.
// navDir is the nav file's directory, relative to the OPF folder.
func epubParseNavXHTML(zr *zip.Reader, zipPath, navDir string) (map[string]string, error) {
	data, err := epubReadZipEntry(zr, zipPath)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	titles := map[string]string{}
	doc.Find("nav a[href]").Each(func(_ int, a *goquery.Selection) {
		href, _ := a.Attr("href")
		href = epubStripFragment(href)
		if href == "" {
			return
		}
		resolved := epubResolveHref(navDir, href)
		title := strings.TrimSpace(a.Text())
		if title != "" && resolved != "" {
			Goose.Logf(5, "epubParseNavXHTML: %s -> %q\n", resolved, title)
			titles[resolved] = title
		}
	})

	return titles, nil
}

// epubParseNCX parses an epub2 NCX file and returns a map from
// OPF-relative href (without fragment) to display title.
// ncxDir is the NCX file's directory, relative to the OPF folder.
func epubParseNCX(zr *zip.Reader, zipPath, ncxDir string) (map[string]string, error) {
	data, err := epubReadZipEntry(zr, zipPath)
	if err != nil {
		return nil, err
	}
	var ncx epubNCX
	if err := xml.Unmarshal(data, &ncx); err != nil {
		return nil, err
	}
	titles := map[string]string{}
	epubCollectNavPoints(ncx.Points, ncxDir, titles)
	return titles, nil
}

func epubCollectNavPoints(points []epubNavPoint, dir string, titles map[string]string) {
	for _, p := range points {
		src := epubStripFragment(p.Content.Src)
		if src != "" {
			resolved := epubResolveHref(dir, src)
			title := strings.TrimSpace(p.Label)
			if title != "" && resolved != "" {
				Goose.Logf(5, "epubCollectNavPoints: %s -> %q\n", resolved, title)
				titles[resolved] = title
			}
		}
		epubCollectNavPoints(p.Points, dir, titles)
	}
}

// epubStripFragment removes the URL fragment (#...) from href.
func epubStripFragment(href string) string {
	if i := strings.Index(href, "#"); i >= 0 {
		return href[:i]
	}
	return href
}

// epubResolveHref resolves href relative to dir (both relative to the OPF
// folder) and returns a cleaned OPF-relative path.
func epubResolveHref(dir, href string) string {
	if path.IsAbs(href) {
		return path.Clean(href[1:])
	}
	if dir == "" || dir == "." {
		return path.Clean(href)
	}
	return path.Clean(dir + "/" + href)
}

// epubZipPath builds the full zip-entry path from the OPF root folder and
// an OPF-relative href.
func epubZipPath(folder, href string) string {
	if folder == "" {
		return href
	}
	return folder + "/" + href
}
