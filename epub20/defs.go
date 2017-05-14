package epub20

import (
   "io"
   "archive/zip"
   "github.com/luisfurquim/ugarit"
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


//Ncx OPS/toc.ncx
type Ncx struct {
   XMLName     struct{}   `xml:"ncx"`
   Version     string     `xml:"version,attr"`
   Langattr    string     `xml:"xml:lang,attr"`
   Xmlns       string     `xml:"xmlns,attr"`
   Metatag   []Metatag    `xml:"head>meta"`
   Title       string     `xml:"docTitle>text,omitempty"`
   Author      string     `xml:"docAuthor>text,omitempty"`
   Points    []NavPoint   `xml:"navMap>navPoint"`
   Pages     []PageTarget `xml:"pageList,omitempty"`
}

//NavPoint nav point
type NavPoint struct {
   Id        string   `xml:"id,attr"`
   PlayOrder string   `xml:"playOrder,attr"`
   Label     string   `xml:"navLabel>text"`
   Content   Content  `xml:"content"`
   Points  []NavPoint `xml:"navPoint"`
}

//Content nav-point content
type Content struct {
   Src string `xml:"src,attr"`
}


type PageTarget struct {
   XMLName struct{}   `xml:"pageTarget"`
   Id      string     `xml:"id,attr"`
   Type    string     `xml:"type,attr"`
   Value   string     `xml:"value,attr"`
   Label   string     `xml:"navLabel>text"`
   Content Content    `xml:"content"`
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
   RootFolder string
}

//Package content.opf
type Package struct {
   XMLName    struct{}   `xml:"package"`
   Version    string     `xml:"version,attr"`
   Xmlns      string     `xml:"xmlns,attr"`
   UID        string     `xml:"unique-identifier,attr"`
   Metadata   Metadata   `xml:"metadata"`
   Manifest []Manifest   `xml:"manifest>item"`
   Spine      Spine      `xml:"spine"`
   Guide      Guide      `xml:"guide,omitempty"`
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
   Langattr string `xml:"xml:lang,attr,omitempty"`
   Content  string `xml:"content,attr,omitempty"`
   Data     string `xml:",chardata"`
}

//Manifest manifest
type Manifest struct {
   ID            string `xml:"id,attr,omitempty"`
   Href          string `xml:"href,attr"`
   MediaType     string `xml:"media-type,attr"`
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

type Guide struct {
   Reference []Reference `xml:"reference,omitempty"`
}

type Reference struct {
   Href      string `xml:"href,attr"`
   Type      string `xml:"type,attr,omitempty"`
   Title     string `xml:"title,attr,omitempty"`
}

type IndexGenerator struct {
   doc Ncx
   curr *[]NavPoint
}


type Store struct {
   w io.Writer
}

