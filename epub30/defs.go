package epub30

import (
   "io"
   "archive/zip"
   "golang.org/x/net/html"
   "github.com/luisfurquim/ugarit"
   "github.com/PuerkitoBio/goquery"
)

type EPubOptions struct {
   Prop int
   TOCTitle string
   TOC ugarit.TOCRef
   TOCItemTitle string
}

type IndexOptions struct {
   IndexGenerator ugarit.IndexGenerator
   Id string
   Prop int
}

type TOCContent struct {
   ndx int
   Title string
   manifest []Manifest
   index []*TOCContent
   ref string
   subSection ugarit.SectionStyle
}

// Book epub book
type Book struct {
   Package       Package       `xml:"package"`
   cwd string
   zfd *zip.Writer
   fd  io.WriteCloser
   fid int
   index []*TOCContent
   subSection ugarit.SectionStyle
}

//Package content.opf
type Package struct {
   XMLName  struct{}   `xml:"package"`
   Version  string     `xml:"version,attr"`
   Langattr string     `xml:"xml:lang,attr"`
   Xmlns    string     `xml:"xmlns,attr"`
   UID      string     `xml:"unique-identifier,attr"`
   Metadata Metadata   `xml:"metadata"`
   Manifest []Manifest `xml:"manifest>item"`
   Spine    Spine      `xml:"spine"`
}

//Metadata metadata
type Metadata struct {
   Xmlns         string     `xml:"xmlns:dc,attr"`
   Title       []string     `xml:"dc:title"`
   Language    []string     `xml:"dc:language"`
   Identifier  []Identifier `xml:"dc:identifier"`
   Creator     []Author     `xml:"dc:creator"`
   Publisher   []string     `xml:"dc:publisher"`
   Date        []Date       `xml:"dc:date"`
   Signature    *Signature  `xml:"link,omitempty"`
   Metatag     []Metatag    `xml:"meta"`
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
   Rel     string `xml:"rel,attr"`
   Href    string `xml:"href,attr"`
}

// Metatag
type Metatag struct {
   Name     string `xml:"name,attr,omitempty"`
   Property string `xml:"property,attr,omitempty"`
   Content  string `xml:"content,attr,omitempty"`
   Data     string `xml:",chardata"`
}

//Manifest manifest
type Manifest struct {
   ID            string `xml:"id,attr,omitempty"`
   Href          string `xml:"href,attr"`
   MediaType     string `xml:"media-type,attr"`
//   MediaFallback string `xml:"media-fallback,attr"`
   Properties    string `xml:"properties,attr,omitempty"`
   MediaOverlay  string `xml:"media-overlay,attr,omitempty"`
}

// Spine spine
type Spine struct {
   ID                string    `xml:"id,attr,omitempty"`
   Toc               string    `xml:"toc,attr,omitempty"`
   PageProgression   string    `xml:"page-progression-direction,attr,omitempty"`
   Itemref         []SpineItem `xml:"itemref"`
}

// SpineItem spine item
type SpineItem struct {
   IDref      string `xml:"idref,attr"`
   ID         string `xml:"id,attr,omitempty"`
   Linear     string `xml:"linear,attr,omitempty"`
   Properties string `xml:"properties,attr,omitempty"`
}

type IndexGenerator struct {
   doc *goquery.Document
   curr *html.Node
}


type Store struct {
   w io.Writer
}

