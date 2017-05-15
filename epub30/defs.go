package epub30

import (
	"archive/zip"
	"github.com/PuerkitoBio/goquery"
	"github.com/luisfurquim/ugarit"
	"golang.org/x/net/html"
	"io"
)

type EPubOptions struct {
	Prop         int
	TOCTitle     string
	TOC          ugarit.TOCRef
	TOCItemTitle string
}

type IndexOptions struct {
	IndexGenerator ugarit.IndexGenerator
	Id             string
	Prop           int
}

type TOCContent struct {
	ndx        int
	Title      string
	manifest   []Manifest
	index      []*TOCContent
	ref        string
	subSection ugarit.SectionStyle
}

// Book epub book
type Book struct {
	Package    Package `xml:"package"`
	cwd        string
	zfd        *zip.Writer
	fd         io.WriteCloser
	fid        int
	index      []*TOCContent
	subSection ugarit.SectionStyle
	RootFolder string
}

//Package content.opf
type Package struct {
	XMLName      struct{}   `xml:"package"`
	Version      string     `xml:"version,attr"`
	Langattr     string     `xml:"xml:lang,attr"`
	Xmlns        string     `xml:"xmlns,attr"`
	XmlnsDc      string     `xml:"xmlns:dc,attr"`
	XmlnsDcterms string     `xml:"xmlns:dcterms,attr"`
	Prefix       string     `xml:"prefix,attr"`
	UID          string     `xml:"unique-identifier,attr"`
	Metadata     Metadata   `xml:"metadata"`
	Manifest     []Manifest `xml:"manifest>item"`
	Spine        Spine      `xml:"spine"`
	Guide        Guide      `xml:"guide,omitempty"`
}

//Metadata metadata
type Metadata struct {
	Xmlns      string       `xml:"xmlns:dc,attr"`
	Title      []string     `xml:"dc:title"`
	Language   []string     `xml:"dc:language"`
	Identifier []Identifier `xml:"dc:identifier"`
	Creator    []Author     `xml:"dc:creator"`
	Publisher  []string     `xml:"dc:publisher"`
	Date       []Date       `xml:"dc:date"`
	Signature  *Signature   `xml:"link,omitempty"`
	Metatag    []Metatag    `xml:"meta"`
}

// Identifier
type Identifier struct {
	Data   string `xml:",chardata"`
	ID     string `xml:"id,attr,omitempty"`
	Scheme string `xml:"scheme,attr,omitempty"`
}

// Author
type Author struct {
	Role   string `xml:"role,attr,omitempty"`
	FileAs string `xml:"file-as,attr,omitempty"`
	Data   string `xml:",chardata"`
}

// Date
type Date struct {
	Event string `xml:"event,attr,omitempty"`
	Data  string `xml:",chardata"`
}

// Signature pathname
type Signature struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// Metatag
type Metatag struct {
	Name     string `xml:"name,attr,omitempty"`
	Langattr string `xml:"xml:lang,attr,omitempty"`
	Property string `xml:"property,attr,omitempty"`
	Content  string `xml:"content,attr,omitempty"`
	Data     string `xml:",chardata"`
}

//Manifest manifest
type Manifest struct {
	ID        string `xml:"id,attr,omitempty"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
	//   MediaFallback string `xml:"media-fallback,attr"`
	Properties   string `xml:"properties,attr,omitempty"`
	MediaOverlay string `xml:"media-overlay,attr,omitempty"`
}

// Spine spine
type Spine struct {
	ID              string      `xml:"id,attr,omitempty"`
	Toc             string      `xml:"toc,attr,omitempty"`
	PageProgression string      `xml:"page-progression-direction,attr,omitempty"`
	Itemref         []SpineItem `xml:"itemref"`
}

// SpineItem spine item
type SpineItem struct {
	IDref      string `xml:"idref,attr"`
	ID         string `xml:"id,attr,omitempty"`
	Linear     string `xml:"linear,attr,omitempty"`
	Properties string `xml:"properties,attr,omitempty"`
}

type Guide struct {
	Reference []Reference `xml:"reference,omitempty"`
}

type Reference struct {
	Href  string `xml:"href,attr"`
	Type  string `xml:"type,attr,omitempty"`
	Title string `xml:"title,attr,omitempty"`
}

type IndexGenerator struct {
	doc  *goquery.Document
	curr *html.Node
}

type Store struct {
	w io.Writer
}

type Versioner struct{}

const (
	prop_Min int = iota

	Prop_Mathml          // mathml
	Prop_RemoteResources // remote-resources
	Prop_Scripted        // scripted
	Prop_Svg             // svg

	prop_Max
	prop_CoverImage // cover-image
	prop_Nav        // nav
)

var prop []string = []string{"", "mathml", "remote-resources", "scripted", "svg", "", "cover-image", "nav"}
