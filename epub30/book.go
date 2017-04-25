package epub30

import (
   "io"
   "fmt"
   "strings"
//   "reflect"
   "archive/zip"
   "encoding/xml"
//   "compress/flate"
   "golang.org/x/net/html"
   "golang.org/x/net/html/atom"
   "github.com/luisfurquim/ugarit"
   "github.com/PuerkitoBio/goquery"
)


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

   b.index = make([]*TOCContent,0,4)
   b.cwd = "/"
   b.fd  = target
   b.zfd = zip.NewWriter(target)


   header := &zip.FileHeader{
      Name:   "mimetype",
      Method: zip.Store,
   }
   f, err := b.zfd.CreateHeader(header)
   if err != nil {
       return nil, err
   }
   _, err = f.Write([]byte("application/epub+zip"))
   if err != nil {
       return nil, err
   }

   f, err = b.zfd.Create("META-INF/container.xml")
   if err != nil {
       return nil, err
   }
   _, err = f.Write([]byte(`<?xml version="1.0"?><container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles><rootfile full-path="EPUB/root.opf" media-type="application/oebps-package+xml" /></rootfiles></container>`))
   if err != nil {
       return nil, err
   }

   b.Package = Package{
      Version: "3.0",
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

   b.subSection = ugarit.NewArabicNumbering("Chapter",true)

   return &b, nil
}

func (b *Book) AddPage(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, ugarit.TOCRef, error) {
   var w io.Writer
   var err error
   var opt *EPubOptions
   var pos int
   var tc *TOCContent

   pos = len(b.Package.Manifest)

   id, w, err = b.AddFile(path, mimetype, src, id, options)
   if err != nil {
      return "", w, nil, err
   }

   fmt.Printf("OPTIONS: %#v\n",options)

   if options != nil {
      switch options.(type) {
         case *EPubOptions:
            opt = options.(*EPubOptions)
            if opt.TOCTitle != "" {
               if opt.TOCItemTitle == "" {
                  return "", w, nil, ugarit.ErrorTOCItemTitleNotFound
               }

               tc = &TOCContent{
                  ndx: pos,
                  Title: opt.TOCTitle,
                  index: make([]*TOCContent,1,4),
                  subSection: ugarit.NewArabicNumbering("Chapter",true),
               }

               if opt.TOC != nil {
                  switch opt.TOC.(type) {
                     case *TOCContent:
                        opt.TOC.(*TOCContent).index = append(opt.TOC.(*TOCContent).index,tc)
                     default:
                        return "", w, nil, ugarit.ErrorInvalidOptionType
                  }
               } else {
                  b.index = append(b.index,tc)
               }

               tc.index = append(tc.index,&TOCContent{
                  ndx:pos,
                  Title: opt.TOCItemTitle,
                  index: make([]*TOCContent,0,4),
                  subSection: ugarit.NewArabicNumbering("Chapter",true),
               })

               fmt.Printf("TOC: %#v\n",b.index)
            } else {
               tc = &TOCContent{
                  ndx:pos,
                  Title: opt.TOCItemTitle,
                  index: make([]*TOCContent,0,4),
                  subSection: ugarit.NewArabicNumbering("Chapter",true),
               }

               if opt.TOC != nil {
                  switch opt.TOC.(type) {
                     case *TOCContent:
                        opt.TOC.(*TOCContent).index = append(opt.TOC.(*TOCContent).index,tc)
                     default:
                        return "", w, nil, ugarit.ErrorInvalidOptionType
                  }
               } else {
                  b.index = append(b.index,tc)
               }
            }
         default:
            return "", nil, nil, ugarit.ErrorInvalidOptionType

      }
   }


   b.Package.Spine.Itemref = append(b.Package.Spine.Itemref,SpineItem{IDref: id})

   return id, w, tc, nil
}


func (b *Book) AddTOC(gen ugarit.IndexGenerator, id string) (string, error) {
   var err error
   var r io.Reader


   for _, ndx := range b.index {
      txt := &html.Node{
         Type: html.TextNode,
         Data: ndx.Title,
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

   id, _, err = b.addFile(gen.GetPathName(), gen.GetMimeType(), r, id, &EPubOptions{}, "nav")

   return id, err
}

func (b *Book) TOCChild(n int) ugarit.TOC {
   if n < len(b.index) {
      return b.index[n]
   }

   return nil
}


func (b *Book) TOCLen() int {
   return len(b.index)
}


func (b *Book) SubSectionStyle(sty ugarit.SectionStyle) {
   b.subSection = sty
}


func (tc *TOCContent) TOCChild(n int) ugarit.TOC {
   if n < len(tc.index) {
      return tc.index[n]
   }

   return nil
}

func (tc *TOCContent) TOCLen() int {
   return len(tc.index)
}

func (tc *TOCContent) ItemRef() string{
   return tc.ref
}

func (tc *TOCContent) ItemTitle() string {
   return tc.Title
}

func (tc *TOCContent) ContentRef() string {
   return tc.manifest[tc.ndx].Href
}

func (tc *TOCContent) SubSectionStyle(sty ugarit.SectionStyle) {
   tc.subSection = sty
}








func (b *Book) AddFile(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, error) {
   var opt *EPubOptions
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
         case *EPubOptions:
            opt = options.(*EPubOptions)
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

func (b *Book) addFile(path string, mimetype string, src io.Reader, id string, opt *EPubOptions, optProp string) (string, io.Writer, error) {
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
      Properties: optProp,
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
   return "application/xhtml+xml"
}

// GetPathName returns the relative pathname of the TOC file
func (gen IndexGenerator) GetPathName() string {
   return "index.xhtml"
}

