package epub201

import (
   "io"
   "fmt"
   "strings"
//   "reflect"
   "archive/zip"
   "encoding/xml"
   "compress/flate"
   "golang.org/x/net/html"
   "golang.org/x/net/html/atom"
   "github.com/luisfurquim/ugarit"
   "github.com/PuerkitoBio/goquery"
)

type BookOptions struct {
   Prop int
   TOC string
}

type IndexOptions struct {
   IndexGenerator ugarit.IndexGenerator
   Id string
   Prop int
}

type TOCContent struct {
   ndx int
   title string
}


//Ncx OPS/toc.ncx
type Ncx struct {
   XMLName     struct{}   `xml:"ncx"`
   Version     string     `xml:"version,attr"`
   Langattr    string     `xml:"xml:lang,attr"`
   Xmlns       string     `xml:"xmlns,attr"`
   Dtbuid      Dtbuid     `xml:"head>meta,omitempty"`
   Title       string     `xml:"docTitle>text,omitempty"`
   Author      string     `xml:"docAuthor>text,omitempty"`
   Points    []NavPoint   `xml:"navMap>navPoint"`
   Pages     []PageTarget `xml:"pageList>pageTarget,omitempty"`
}

type Dtbuid struct {
   Name    string `xml:"name,attr"`
   Content string `xml:"content,attr"`
}

//NavPoint nav point
type NavPoint struct {
   Label     string   `xml:"navLabel>text"`
   Content   Content  `xml:"content"`
   Points  []NavPoint `xml:"navPoint"`
}

//Content nav-point content
type Content struct {
   Src string `xml:"src,attr"`
}

type PageTarget struct {
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
   index []TOCContent
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

func (w Store) Write(p []byte) (n int, err error) {
   return w.w.Write(p)
}

func (w Store) Close() error {
   return nil
}

const (
   prop_Min int = iota

   Prop_Mathml // mathml
   Prop_RemoteResources // remote-resources
   Prop_Scripted // scripted
   Prop_Svg // svg

   prop_Max
   prop_CoverImage // cover-image
   prop_Nav // nav
)


var nav string = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="en"
   lang="en">
   <head>
      <title class="title"></title>
   </head>
   <body>
      <h1 class="title"></h1>
      <nav epub:type="toc" id="toc">
         <h2 id="toc"></h2>
         <ol id="level0"></ol>
      </nav>
   </body>
</html>`



// Create a blank EPub Book
func New(
      target            io.WriteCloser,
      Title           []string,
      Language        []string,
      Identifier      []Identifier,
      Creator         []Author,
      Publisher       []string,
      Modified          string,
      Date            []Date,
      signature         Signature,
      Metatag         []Metatag,
      PageProgression   string) (*Book, error) {
   var b Book
   var Sig *Signature

   b.index = []TOCContent{}
   b.cwd = "/"
   b.fd  = target
   b.zfd = zip.NewWriter(target)

   b.zfd.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
      return Store{w: out}, nil
   })

   f, err := b.zfd.Create("mimetype")
   if err != nil {
       return nil, err
   }
   _, err = f.Write([]byte("application/epub+zip"))
   if err != nil {
       return nil, err
   }

   b.zfd.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
      return flate.NewWriter(out, flate.BestCompression)
   })

   f, err = b.zfd.Create("META-INF/container.xml")
   if err != nil {
       return nil, err
   }
   _, err = f.Write([]byte(`<?xml version="1.0"?><container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="EPUB/root.opf" media-type="application/oebps-package+xml" /></rootfiles></container>`))
   if err != nil {
       return nil, err
   }

/*
   Sig = &Signature{
      Rel: "xml-signature",
      Href: signature.Href,
   }

*/

   b.Package = Package{
      Version: "2.0.1",
      Xmlns:   "http://www.idpf.org/2007/opf",
      UID:     "pub-id",
      Metadata: Metadata{
         Xmlns:      "http://purl.org/dc/elements/1.1/",
         Title:      Title,
         Language:   Language,
         Identifier: Identifier,
         Creator:    Creator,
         Publisher:  Publisher,
         Date:       Date,
         Signature:  Sig,
         Metatag:       Metatag,
      },
      Manifest: []Manifest{
      },
      Spine:    Spine{
         Itemref:         []SpineItem{},
      },
   }

   if PageProgression != "" {
      b.Package.Spine.PageProgression = PageProgression
   }

   if len(Language) > 0 {
      b.Package.Langattr = Language[0]
   } else {
      b.Package.Langattr = "en"
   }

   return &b, nil
}

func (b *Book) AddPage(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, error) {
   var w io.Writer
   var err error
   var opt *BookOptions
   var pos int

   pos = len(b.Package.Manifest)

   id, w, err = b.AddFile(path, mimetype, src, id, options)
   if err != nil {
      return "", w, err
   }

   fmt.Printf("OPTIONS: %#v\n",options)

   if options != nil {
      switch options.(type) {
         case *BookOptions:
            opt = options.(*BookOptions)
            if opt.TOC != "" {
               b.index = append(b.index,TOCContent{
                  ndx:pos,
                  title: opt.TOC,
               })
               fmt.Printf("TOC: %#v\n",b)
            }
         default:
            return "", nil, ugarit.ErrorInvalidOptionType

      }
   }


   b.Package.Spine.Itemref = append(b.Package.Spine.Itemref,SpineItem{IDref: id})

   return id, w, nil
}


func (b *Book) AddTOC(gen ugarit.IndexGenerator, id string) (string, error) {
   var err error
   var r io.Reader


   for _, ndx := range b.index {
      txt := &html.Node{
         Type: html.TextNode,
         Data: ndx.title,
      }
      err = gen.AddItem(&html.Node{
         FirstChild: txt,
         LastChild:  txt,
         Type:       html.ElementNode,
         DataAtom:   atom.A,
         Data:       "A",
         Attr: []html.Attribute{html.Attribute{
            Key: "href",
            Val: b.Package.Manifest[ndx.ndx].Href,
         }},
      })
   }

   r, err = gen.GetDocument()
   if err != nil {
      return "", err
   }

   if id == "" {
      id = "nncx"
   }

   id, _, err = b.addFile(gen.GetPathName(), gen.GetMimeType(), r, id, &BookOptions{}, "")

   return id, err
}


func (b *Book) AddFile(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, error) {
   var opt *BookOptions
   var optProp string

   if len(path) == 0 {
      return "", nil, ugarit.ErrorInvalidPathname
   }

   if path == "/" {
      return "", nil, ugarit.ErrorInvalidPathname
   }

   if path == "/index.html" || path == "index.html"{
      return "", nil, ugarit.ErrorInvalidPathname
   }

   if options != nil {
      switch options.(type) {
         case *BookOptions:
            opt = options.(*BookOptions)
         default:
            return "", nil, ugarit.ErrorInvalidOptionType
      }
   }

   if opt.Prop>prop_Min && opt.Prop<prop_Max {
      switch opt.Prop {
         case Prop_Mathml:
            optProp = "mathml"
         case Prop_RemoteResources:
            optProp = "remote-resources"
         case Prop_Scripted:
            optProp = "scripted"
      }

   }

   return b.addFile(path, mimetype, src, id, opt, optProp)
}

func (b *Book) addFile(path string, mimetype string, src io.Reader, id string, opt *BookOptions) (string, io.Writer, error) {
   var err error
   var w io.Writer

   if b.Package.Manifest == nil {
      b.Package.Manifest = []Manifest{}
   }

   if id == "" {
      id = fmt.Sprintf("pg%d",b.fid)
      b.fid++
   }

   b.Package.Manifest = append(b.Package.Manifest,Manifest{
      ID: id,
      Href: path,
      MediaType: mimetype,
   })

   if path[0] == '/' {
      path = "EPUB" + path
   } else {
      path = "EPUB/" + path
   }

   w, err = b.zfd.Create(path)
   if err != nil {
      return "", nil, err
   }

   if src != nil {
      io.Copy(w,src)
      return id, nil, nil
   }

   return id, w, nil
}


func (b *Book) Close() error {
   var enc *xml.Encoder

   f, err := b.zfd.Create("EPUB/root.opf")
   if err != nil {
       return err
   }

   f.Write([]byte(xml.Header))

   enc = xml.NewEncoder(f)
   err = enc.Encode(b.Package)
   if err != nil {
       return err
   }


   err = b.zfd.Close()
   if err != nil {
      return err
   }

   return b.fd.Close()
}


func NewIndexGenerator() (*IndexGenerator, error) {
   var ig IndexGenerator
   var err error

   ig.doc, err = goquery.NewDocumentFromReader(strings.NewReader(nav))
   if err != nil {
      return nil, err
   }

   ig.curr = ig.doc.Find("OL").Nodes[0]
   return &ig, nil
}

func (gen IndexGenerator) AddItem(item *html.Node) error {
   gen.curr.AppendChild(
      &html.Node{
         FirstChild: item,
         LastChild:  item,
         Type:       html.ElementNode,
         DataAtom:   atom.Li,
         Data:       "LI",
      })
   return nil
}

func (gen IndexGenerator) GetDocument() (io.Reader, error) {
   var doc string
   var err error

   doc, err = gen.doc.Html()
   if err != nil {
      return nil, err
   }

   doc = doc[59:]

   return strings.NewReader(doc), nil
}



// GetMimeType returns the mimetype of the TOC file.
func (gen IndexGenerator) GetMimeType() string {
   return "application/x-dtbncx+xml"
}

// GetPathName returns the relative pathname of the TOC file
func (gen IndexGenerator) GetPathName() string {
   return "index.ncx"
}

