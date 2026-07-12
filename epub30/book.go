package epub30

import (
//   "os"
   "io"
   "fmt"
   "time"
   "sort"
   "regexp"
   "strings"
   "io/ioutil"
   "archive/zip"
   "encoding/xml"
   "path/filepath"
   "golang.org/x/net/html"
   "golang.org/x/net/html/atom"
   "github.com/PuerkitoBio/goquery"
   "github.com/luisfurquim/ugarit"
)

// Create a blank EPub Book
//
// Parameters:
//
// @Target -- where to save the epub contents
//
// @Title -- eBook title
//
// @Language -- eBook language
//
// @identifier -- eBook ID
//
// @Creator -- eBook author(s)
//
// @Publisher -- eBook Publisher(s)
//
// @Date -- Date published
//
// @signature -- Ignored for now :(
//
// @metatag -- Any epub metatags go here
//
// @PageProgression -- PageProgression use "ltr" (left-to-right) or "rtl" (right-to-left)
//
// @bookversion -- provide a versioner interface or use the package provided one which just generate a version string using the current time
func New(
   target io.WriteCloser, // where to save the epub contents
   title []string, // eBook title
   language []string, // eBook language
   identifier []string, // eBook ID
   creator []Author, // eBook author(s)
   publisher []string, // eBook Publisher(s)
   date []Date, // Date published
   signature Signature, // Ignored for now :(
   metatag []Metatag, // Any epub metatags go here
   pageProgression string, // PageProgression use "ltr" (left-to-right) or "rtl" (right-to-left)
   // provide a versioner interface or use the package provided one
   // which just generate a version string using the current time
   bookversion ugarit.Versioner) (*Book, error) {
   var b Book
   var sig *Signature
   var ids []Identifier

   b.RootFolder = "OEBPS"

   b.index = make(TOC, 0, 4)
   b.cwd = "/"
   b.fd = target
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
   _, err = f.Write([]byte(fmt.Sprintf(rootfolder, b.RootFolder)))
   if err != nil {
      return nil, err
   }

   ids = make([]Identifier, len(identifier))
   for i, id := range identifier {
      ids[i] = Identifier{
         Data: id,
         ID:   "pub-id",
      }
   }

   if bookversion != nil {
      metatag = append(metatag, Metatag{
         Property: "ibooks:version",
         Data:     bookversion.Next(),
      })
   }

   b.Package = Package{
      Version:      "3.0",
      Xmlns:        "http://www.idpf.org/2007/opf",
      XmlnsDc:      "http://purl.org/dc/elements/1.1/",
      XmlnsDcterms: "http://purl.org/dc/terms/",
      Prefix:       "ibooks: http://vocabulary.itunes.apple.com/rdf/ibooks/vocabulary-extensions-1.0/",
      UID:          "pub-id",
      Metadata: Metadata{
         Xmlns:      "http://purl.org/dc/elements/1.1/",
         Title:      title,
         Language:   language,
         Identifier: ids,
         Creator:    creator,
         Publisher:  publisher,
         Date:       date,
         Signature:  sig,
         Metatag:    metatag,
      },
      //    Spine: Spine{
      //       Itemref: []SpineItem{},
      //    },
   }

   if pageProgression != "" {
      b.Package.Spine.PageProgression = pageProgression
   }

   b.Package.Langattr = "en"
   if len(language) > 0 {
      b.Package.Langattr = language[0]
      // } else {
      //    b.Package.Langattr = "en"
   }

   b.subSection = ugarit.NewArabicNumbering("Chapter", true)

   b.ManifIndex = map[string]string{}

   return &b, nil
}

// AddPage saves the page to the E-Book and adds an entry in the Spine pointing to it.
// It calls AddFile. So, if you use this method, you don't need to call AddFile to save it,
// otherwise it will be stored twice in the E-Book file.
// AddPage creates/truncates the file specified by path inside the epub file,
// register as having the provided mimetype and making it to figure in the book index.
// If src is nil, it returns a valid io.Writer.
// If src is not nil, it copies its content to the file and return a nil io.Writer.
// If path collides with a reserved path, it returns an error.
// If the page is to be added to the TOC, provide an EPubOptions object containing
// a TOCTitle and a TOCItemTitle; OR a TOCContent object.
// TOC support is still alpha code and will be improved in the future
func (b *Book) AddPage(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, ugarit.TOCRef, error) {
   var w io.Writer
   var err error
   var opt *EPubOptions
   var pos int
   var tc *TOCContent
   var doc *goquery.Document
   var shtml string

   if options != nil {
      switch options.(type) {
      case *EPubOptions:
         opt = options.(*EPubOptions)
         if src != nil && opt.FilterHTML != nil {
            doc, err = goquery.NewDocumentFromReader(src)
            if err != nil {
               return "", nil, nil, err
            }
            opt.Prop = append(opt.Prop,opt.FilterHTML(doc.Nodes)...)
            shtml, err = doc.Html()
            if err != nil {
               return "", nil, nil, err
            }
            src = strings.NewReader("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE html>\n" + shtml)
         }

         if src != nil && opt.Width != 0 && opt.Height != 0 {
            doc, err = goquery.NewDocumentFromReader(src)
            if err != nil {
               return "", nil, nil, err
            }
            doc.Find("HEAD").Each(func(_ int, s *goquery.Selection) {
					s.AppendHtml(fmt.Sprintf(`<meta name="viewport" content="width=%d, height=%d"></meta>`, opt.Width, opt.Height))
				})
            shtml, err = doc.Html()
            if err != nil {
               return "", nil, nil, err
            }
            src = strings.NewReader("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE html>\n" + shtml)
			}

      default:
         return "", nil, nil, ugarit.ErrorInvalidOptionType
      }
   }

   pos = len(b.Package.Manifest)

   id, w, err = b.AddFile(path, mimetype, src, id, opt)
   if err != nil {
      return "", w, nil, err
   }

   //   fmt.Printf("OPTIONS: %#v\n",options)

   if opt != nil {
      if opt.TOCItemTitle == "" {
         return "", w, nil, ugarit.ErrorTOCItemTitleNotFound
      }

      if opt.TOCTitle != "" {
         tc = &TOCContent{
            ndx:        pos,
            Title:      opt.TOCTitle,
            index:      make(TOC, 1, 4),
            subSection: ugarit.NewArabicNumbering("Chapter", true),
         }

         if opt.TOC != nil {
            switch opt.TOC.(type) {
            case *TOCContent:
               opt.TOC.(*TOCContent).index = append(opt.TOC.(*TOCContent).index, tc)
            default:
               return "", w, nil, ugarit.ErrorInvalidOptionType
            }
         } else {
            b.index = append(b.index, tc)
         }

         tc.index[0] = &TOCContent{
            ndx:        pos,
            Title:      opt.TOCItemTitle,
            index:      make(TOC, 0, 4),
            subSection: ugarit.NewArabicNumbering("Chapter", true),
         }

//         tc = tc.index[0]

//            fmt.Printf("TOC: %#v\n", b.index)
      } else {
         tc = &TOCContent{
            ndx:        pos,
            Title:      opt.TOCItemTitle,
            index:      make(TOC, 0, 4),
            subSection: ugarit.NewArabicNumbering("Chapter", true),
         }

         if opt.TOC != nil {
//               fmt.Printf("TOC before: %s\n", opt.TOC.(*TOCContent))
            switch opt.TOC.(type) {
            case *TOCContent:
               opt.TOC.(*TOCContent).index = append(opt.TOC.(*TOCContent).index, tc)
//                  fmt.Printf("SUBTOC: %s\n", opt.TOC.(*TOCContent))
            default:
               return "", w, nil, ugarit.ErrorInvalidOptionType
            }
         } else {
            b.index = append(b.index, tc)
         }
//            fmt.Printf("TOC: %s\n", b.index)
      }

   }

   b.Package.Spine.Itemref = append(b.Package.Spine.Itemref, SpineItem{IDref: id})

   return id, w, tc, nil
}

// AddCover saves the Cover in the E-Book. Use it before saving any content to the E-Book.
// AddCover creates/truncates the file specified by path inside the epub file,
// registers as having the provided mimetype and makes it be the book cover.
// If src is nil, it returns a valid io.Writer.
// If src is not nil, it copies its contents to the file and returns a nil io.Writer.
// If path collides with a reserved path, it returns an error.
func (b *Book) AddCover(path string, mimetype string, src io.Reader, options interface{}) (string, io.Writer, error) {
   var w io.Writer
   var err error
   var opt, opt2 *EPubOptions
   var id string
   var svgcontent []byte
   var pathhtml string
   var extension string
//   var fileprop []string

   if options != nil {
      switch options.(type) {
      case *EPubOptions:
         opt = options.(*EPubOptions)
         if opt == nil {
				opt = &EPubOptions{}
			}
      default:
         return "", nil, ugarit.ErrorInvalidOptionType
      }
   } else {
		opt = &EPubOptions{}
	}

   b.Package.Metadata.Metatag = append(b.Package.Metadata.Metatag, Metatag{
      Name:    "cover",
      Content: "cover",
   })

   if mimetype == "image/svg+xml" {
      svgcontent, err = ioutil.ReadAll(src)
      if err != nil {
         return "", w, err
      }

      extension = filepath.Ext(path)
      pathhtml = path[:len(path)-len(extension)] + ".xhtml"
      if opt.Width != 0 && opt.Height != 0 {
			src = strings.NewReader(fmt.Sprintf(svgcoverSized, opt.Width, opt.Height, -opt.Width, -opt.Height, -opt.MarginLeft, -opt.MarginRight, -opt.MarginTop, -opt.MarginBottom, svgcontent))
		} else {
			src = strings.NewReader(fmt.Sprintf(svgcover, svgcontent))
		}

      if opt == nil {
         opt = &EPubOptions{Prop: []int{prop_CoverImage}}
      } else {
         opt.Prop = append(opt.Prop,prop_CoverImage)
      }
//      fileprop = []string{prop[prop_CoverImage]}

   } else if mimetype[:6] == "image/" {
      opt2 = &EPubOptions{Prop: []int{prop_CoverImage}}
      id, w, err = b.addFile(path, mimetype, src, "cover-image", opt2, nil)
//      id, w, err = b.addFile(path, mimetype, src, "cover-image", opt, []string{prop[prop_CoverImage]})
      if err != nil {
         return "", w, err
      }

      mimetype = "application/xhtml+xml"
      if opt.Width != 0 && opt.Height != 0 {
			src = strings.NewReader(fmt.Sprintf(ImgcoverSized, opt.Width, opt.Height, -opt.Width, -opt.Height, -opt.MarginLeft, -opt.MarginRight, -opt.MarginTop, -opt.MarginBottom, path))
		} else {
			src = strings.NewReader(fmt.Sprintf(Imgcover, path))
		}

      extension = filepath.Ext(path)
      pathhtml = path[:len(path)-len(extension)] + ".xhtml"

   } else {
      pathhtml = path
   }

   id, w, err = b.addFile(pathhtml, mimetype, src, "cover", opt, nil)
   if err != nil {
      return "", w, err
   }

   b.Package.Spine.Itemref = append(b.Package.Spine.Itemref, SpineItem{IDref: id, Linear: "yes"})
   b.Package.Guide.Reference = []Reference{Reference{Href: pathhtml, Type: "cover", Title: "Cover"}}
   b.coverPath = pathhtml

   return id, w, nil
}

// mkTOC Construct the Table of Contents
func mkTOC(tc ugarit.TOC, gen ugarit.IndexGenerator, manif []Manifest) error {
   var i int
   var tcont *TOCContent
   var err error
   var txt *html.Node

   for i=0; i<tc.TOCLen(); i++ {
      tcont = tc.TOCChild(i).(*TOCContent)

      txt = &html.Node{
         Type: html.TextNode,
         Data: tcont.Title,
      }
      err = gen.AddItem(&html.Node{
         FirstChild: txt,
         LastChild:  txt,
         Type:       html.ElementNode,
         DataAtom:   atom.A,
         Data:       "a",
         Attr: []html.Attribute{
            html.Attribute{
               Key: "href",
               Val: manif[tcont.ndx].Href,
            },
            html.Attribute{
               Key: "id",
               Val: fmt.Sprintf("pg%d", i+1),
            },
         },
      })
      if err != nil {
         return err
      }

      if len(tcont.index) > 0 {
         gen.AddSection()
         err = mkTOC(tcont, gen, manif)
         gen.EndSection()
         if err != nil {
            return err
         }
      }
   }

   return nil
}

// AddTOC saves the TOC file to the E-Book. Use it just before closing the E-Book.
// AddTOC creates/truncates the TOC file when the book is closed.
// If IndexGenerator != nil,  the generated html nodes are be passed through it before saved in the file.
// If id!="" it sets the TOC id.
func (b *Book) AddTOC(gen ugarit.IndexGenerator, id string) (string, error) {
   var err error
   var r io.Reader
   var path string
   var ref Reference
   var found bool

   if id == "" {
      id = gen.GetId()
   }

   // The cover landmark is only real when AddCover ran; a dangling
   // cover.xhtml reference fails epubcheck on coverless books.
   if b.coverPath != "" {
      if ig, ok := gen.(*IndexGenerator); ok {
         ig.AddLandmark(b.coverPath, "cover", "Cover")
      }
   }

   err = mkTOC(b, gen, b.Package.Manifest)
   if err != nil {
      return id, err
   }

   r, err = gen.GetDocument()
   if err != nil {
      //      fmt.Printf("\nError getting TOC: %s\n\n",err)
      return "", err
   }

   path = gen.GetPathName()
   if gen.GetPropertyValue() == "nav" {
      id, _, err = b.addFile(path, gen.GetMimeType(), r, id, &EPubOptions{Prop: []int{prop_Nav}}, nil)
   } else {
      id, _, err = b.addFile(path, gen.GetMimeType(), r, id, &EPubOptions{}, nil)
   }
   if err == nil {
      // The TOC page reads BEFORE the content, printed-book order. Out of
      // the spine its placement is reader-defined (some append it at the
      // END of the book); first is the only deterministic choice.
      b.Package.Spine.Itemref = append([]SpineItem{{IDref: id}}, b.Package.Spine.Itemref...)
   }
//   id, _, err = b.addFile(path, gen.GetMimeType(), r, id, &EPubOptions{}, []string{gen.GetPropertyValue()})
   //fmt.Printf("\nSaved TOC: %s\n\n", err)

   ref = Reference{Href: path, Type: "toc", Title: "Table of Contents"}

   if b.Package.Guide.Reference == nil {
      b.Package.Guide.Reference = []Reference{ref}
   } else {
      for _, r := range b.Package.Guide.Reference {
         if r.Type == "toc" {
            found = true
            break
         }
      }
      if !found {
         b.Package.Guide.Reference = append(b.Package.Guide.Reference,ref)
      }
   }

   return id, err
}

// TOCChild retrieves the Nth child of the TOC top level
// Compatibility break: it previously was "TOCChild(n int) TOC"
func (b *Book) TOCChild(n int) ugarit.TOCRef {
   if n < len(b.index) {
      return b.index[n]
   }

   return nil
}

// TOCLen retrieves the item count in the TOC top level section
func (b *Book) TOCLen() int {
   return len(b.index)
}

// SubSectionStyle sets the object to style the top level section of the TOC
func (b *Book) SubSectionStyle(sty ugarit.SectionStyle) {
   b.subSection = sty
}

// TOCChild retrieves the Nth child of the TOC item
// Compatibility break: it previously was "TOCChild(n int) TOC"
func (tc TOCContent) TOCChild(n int) ugarit.TOCRef {
   if n < len(tc.index) {
      return tc.index[n]
   }

   return nil
}

// TOCLen retrieves the item count in this TOC section
func (tc TOCContent) TOCLen() int {
   return len(tc.index)
}

// ItemRef retrieves the reference ID of the TOC Item
func (tc TOCContent) ItemRef() string {
   return tc.ref
}

// ItemTitle retrieves the item title (label) as it appears in the TOC
func (tc TOCContent) ItemTitle() string {
   return tc.Title
}

// ContentRef retrieves the path o the file pointed by the TOC entry
func (tc TOCContent) ContentRef() string {
   return tc.manifest[tc.ndx].Href
}

// SubSectionStyle sets the object to style this TOC subsection
func (tc TOCContent) SubSectionStyle(sty ugarit.SectionStyle) {
   tc.subSection = sty
}

// AddFile stores the file in the E-Book. It does not include it in the TOC.
// AddFile creates/truncates the file specified by path,
// register as having the provided mimetype.
// If src is nil, it returns a valid io.Writer.
// If src is not nil, it copies its contents to the file and return a nil io.Writer.
// If path collides with a reserved path, it returns an error.
func (b *Book) AddFile(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, error) {
   var opt *EPubOptions
   var optProp []string

   if len(path) == 0 {
      return "", nil, ugarit.ErrorInvalidPathname
   }

   if path == "/" {
      return "", nil, ugarit.ErrorInvalidPathname
   }

   if path == "/index.html" || path == "index.html" {
      return "", nil, ugarit.ErrorInvalidPathname
   }

   if id == "cover" {
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

   if opt != nil {
      for _, pr := range opt.Prop {
//         if pr > prop_Min && pr < prop_Max {
            optProp = append(optProp,prop[pr])
//         }
      }
   }

   return b.addFile(path, mimetype, src, id, opt, optProp)
}

func (b *Book) AddReference(path string, mimetype string, id string, options interface{}) (string, error) {
   var opt *EPubOptions
   var optProp []string
   var ok bool
   var oldId string

   if oldId, ok = b.ManifIndex[path]; ok {
      return oldId, nil
   }


   if id == "" {
      id = fmt.Sprintf("pg%d", b.fid)
      b.fid++
   }

   if options != nil {
      switch options.(type) {
      case *EPubOptions:
         opt = options.(*EPubOptions)
      default:
         return "", ugarit.ErrorInvalidOptionType
      }
   }

   if opt != nil {
      for _, pr := range opt.Prop {
         optProp = append(optProp,prop[pr])
      }

      if opt.FilterPath != nil {
         path = opt.FilterPath(path)
      }
   } else {
      optProp = []string{}
   }

   b.Package.Manifest = append(b.Package.Manifest, Manifest{
      ID:         id,
      Href:       path,
      MediaType:  mimetype,
      Properties: strings.Join(optProp," "),
   })

   b.ManifIndex[path] = id

   return id, nil
}

func (b *Book) addFile(path string, mimetype string, src io.Reader, id string, opt *EPubOptions, optProp []string) (string, io.Writer, error) {
   var err error
   var ok bool
   var oldId string

   if oldId, ok = b.ManifIndex[path]; ok {
      return oldId, nil, nil
   }

   if id == "" {
      id = fmt.Sprintf("pg%d", b.fid)
      b.fid++
   }

   if opt != nil && opt.FilterPath != nil {
      path = opt.FilterPath(path)
   }

   id, err = b.AddReference(path, mimetype, id, opt)
   if err != nil {
      return id, nil, err
   }

   return b.addfile(path, src, id)
}

func (b *Book) addfile(path string, src io.Reader, id string) (string, io.Writer, error) {
   var err error
   var w io.Writer

   if path[0] == '/' {
      path = b.RootFolder + path
   } else {
      path = b.RootFolder + "/" + path
   }

   w, err = b.zfd.Create(path)
   if err != nil {
      return "", nil, err
   }

   if src != nil {
      io.Copy(w, src)
      return id, nil, nil
   }

   return id, w, nil
}

// Sets spine attributes
func (b *Book) SpineAttr(key, val string) {
   switch key {
   case "id":
      b.Package.Spine.ID = val
   case "toc":
      b.Package.Spine.Toc = val
   case "pageprogression":
      b.Package.Spine.PageProgression = val
   }
}

// Closes and saves the E-Book
func (b *Book) Close() error {
   var enc *xml.Encoder

   b.Package.Metadata.Metatag = append(
      b.Package.Metadata.Metatag,
      Metatag{
         Property: "dcterms:modified",
         Data:     time.Now().Format("2006-01-02T15:04:05Z"),
      })

   f, err := b.zfd.Create(b.RootFolder + "/content.opf")
   if err != nil {
      return err
   }

   f.Write([]byte(xml.Header))


   enc = xml.NewEncoder(f)
   err = enc.Encode(b.Package)
   if err != nil {
      return err
   }

   f, err = b.zfd.Create("META-INF/com.apple.ibooks.display-options.xml")
   if err != nil {
      return err
   }

   f.Write([]byte(iBooksFonts))

   err = b.zfd.Close()
   if err != nil {
      return err
   }

   return b.fd.Close()
}

// Add custom metadata
func (b *Book) AddMetadata(key, val string) {

   b.Package.Metadata.Metatag = append(
      b.Package.Metadata.Metatag,
      Metatag{
         Property: key,
         Data:     val,
      },
   )
}

// Creates an index generator object
func NewIndexGenerator(title ...string) (*IndexGenerator, error) {
   var ig IndexGenerator
   var err error
   var tpl string

   if len(title) > 0 {
		tpl = fmt.Sprintf(nav,title[0])
	} else {
		tpl = fmt.Sprintf(nav,"Document")
	}

   ig.doc, err = goquery.NewDocumentFromReader(strings.NewReader(tpl))
   if err != nil {
      return nil, err
   }

   ig.curr = &[]*html.Node{ig.doc.Find("#TOClevel0").Nodes[0]}

   return &ig, nil
}

// AddLandmark appends one entry to the landmarks nav. The cover landmark
// is added automatically by AddTOC when the book has a cover; callers may
// add others (titlepage, bodymatter, bibliography...).
func (gen *IndexGenerator) AddLandmark(href, epubType, label string) {
   var node *html.Node

   node = &html.Node{
      Type: html.TextNode,
      Data: label,
   }

   node = &html.Node{
      FirstChild: node,
      LastChild:  node,
      Type:       html.ElementNode,
      DataAtom:   atom.A,
      Data:       "a",
      Attr: []html.Attribute{
         html.Attribute{
            Key: "href",
            Val: href,
         },
         html.Attribute{
            Key: "epub:type",
            Val: epubType,
         },
      },
   }

   gen.doc.Find("#lmarks").Nodes[0].AppendChild(
      &html.Node{
         FirstChild: node,
         LastChild:  node,
         Type:       html.ElementNode,
         DataAtom:   atom.Li,
         Data:       "li",
      })
}

//        <li><a href="titlepg.xhtml" epub:type="titlepage">Title Page</a></li>
//        <li><a href="chapter.xhtml" epub:type="bodymatter">Start</a></li>
//        <li><a href="bibliog.xhtml" epub:type="bibliography">Bibliography</a></li>

// AddItem adds a new TOC entry
func (gen *IndexGenerator) AddItem(item *html.Node) error {
   var i int

   gen.id++
   for i=0; i<len(item.Attr); i++ {
      if strings.ToLower(item.Attr[i].Key) == "id" {
         item.Attr[i].Val = fmt.Sprintf("pg%d",gen.id)
         break
      }
   }
   if i == len(item.Attr) {
      item.Attr = append(item.Attr,html.Attribute{
         Key: "id",
         Val: fmt.Sprintf("pg%d",gen.id),
      })
   }

   item.FirstChild.Data = strings.Repeat(" ",gen.level*5) + item.FirstChild.Data;

//   fmt.Printf("Curr: |%d|->%#v\n",len(*gen.curr),gen.curr)
//   fmt.Printf("Curr[-1]: %#v\n",(*gen.curr)[len(*gen.curr)-1])


   (*gen.curr)[len(*gen.curr)-1].AppendChild(&html.Node{
      FirstChild: item,
      LastChild:  item,
      Type:       html.ElementNode,
      DataAtom:   atom.Li,
      Data:       "li",
   })
   return nil
}

// AddSection adds a new section in the current TOC item
func (gen *IndexGenerator) AddSection() error {
   var ol *html.Node

   gen.level++

   ol = &html.Node{
      Type:       html.ElementNode,
      DataAtom:   atom.Ol,
      Data:       "ol",
   }

   if (*gen.curr)[len(*gen.curr)-1].LastChild == nil {
      return ErrorMustAddItemToStartNewSection
   }

   (*gen.curr)[len(*gen.curr)-1].LastChild.AppendChild(ol)

   *gen.curr = append(*gen.curr, ol)

//   fmt.Printf("%#v\n",gen.curr)

   return nil
}

// EndSection finishes the current section and goes up a section level
func (gen *IndexGenerator) EndSection() error {
   if len(*gen.curr) == 1 {
      return ugarit.ErrorAlreadyTopLevel
   }

   gen.level--
   *gen.curr = (*gen.curr)[:len(*gen.curr)-1]

   return nil
}

// GetDocument returns the XHTML content of the TOC file
func (gen *IndexGenerator) GetDocument() (io.Reader, error) {
   var doc string
   var err error

   // An empty <ol> is invalid; a landmarks nav with no entries goes away.
   lmarks := gen.doc.Find("#lmarks")
   if lmarks.Length() > 0 && lmarks.Children().Length() == 0 {
      lmarks.Parent().Remove()
   }

   doc, err = gen.doc.Html()
   if err != nil {
      return nil, err
   }

   doc = doc[59:]

   return strings.NewReader(doc), nil
}

// GetMimeType returns the mimetype of the TOC file.
func (gen *IndexGenerator) GetMimeType() string {
   return "application/xhtml+xml"
}

// GetPathName returns the relative pathname of the TOC file
func (gen *IndexGenerator) GetPathName() string {
   return "index.xhtml"
}

// GetPropertyValue returns the the value to set in the property attribute
// of the TOC OPF entry.
func (gen *IndexGenerator) GetPropertyValue() string {
   return "nav"
}

// GetId returns the ID to use in the TOC OPF entry and the Spine TOC reference.
func (gen *IndexGenerator) GetId() string {
   return "nav"
}

// Simple E-Book versioner
func (v Versioner) Next() string {
   return time.Now().Format("20060102150405")
}



func (tc TOCContent) String() string {
   var s string

   for _, ndx := range tc.index {
      s += ndx.Title + "; "
   }
   return fmt.Sprintf("%s [%s]",tc.Title,s)
}

/*
func (tc *TOCContent) String() string {
   return fmt.Sprintf("%s",*tc)
}
*/

/*
func (tc TOC) String() string {
   var s string
   for i, ndx := range tc {
      s += tc.String() + " // "
   }
   return fmt.Sprintf("%s",*tc)
}
*/


func FilterPathASCDigitDash(s string) string {
   return ASCDigitDash.ReplaceAllString(s,"")
}

func FilterHTMLStd(root []*html.Node) []int {
   var moreOptions []int
   var hasRemoteRef bool
   var elemName string
   var attr string
   var attrList []string
   var doc *goquery.Document
   var attrNode html.Attribute

   if len(root) == 0 {
      return nil
   }

   doc = goquery.NewDocumentFromNode(root[0])

   for elemName, attrList = range allow_attr {
      doc.Find(elemName).Each(func(_ int, sel *goquery.Selection) {
         var i, j int
         for i=0; i<len(sel.Nodes[0].Attr); {
            attrNode = sel.Nodes[0].Attr[i]
            if attrNode.Namespace != "" {
               attr = attrNode.Namespace + ":" + attrNode.Key
            } else {
               attr = attrNode.Key
            }
            if attr == "" {
               sel.Nodes[0].Attr = append(sel.Nodes[0].Attr[:i],sel.Nodes[0].Attr[i+1:]...)
               continue
            }
            j = sort.Search(len(attrList), func(k int) bool { return attrList[k] >= attr })
            if j >= len(attrList) || attrList[j] != attr {
//               fmt.Printf("Removing %s -> %#v from iframe\n", attr, sel.Nodes[0].Attr[i])
               sel.Nodes[0].Attr = append(sel.Nodes[0].Attr[:i],sel.Nodes[0].Attr[i+1:]...)
               continue
            }
            i++
         }
      })
   }

   doc.Find("img[border]").RemoveAttr("border")
   doc.Find("[allowfullscreen]").RemoveAttr("allowfullscreen")
   doc.Find("[frameborder]").RemoveAttr("frameborder")

//   fmt.Printf("\n\ncheck js\n\n\n")
   if doc.Find("script").Length() > 0 {
//      fmt.Printf("\n\nhas js\n\n\n")
      moreOptions = append(moreOptions, Prop_Scripted)
   } else {
//      fmt.Printf("\n\ncheck js href\n\n\n")
      doc.Find("[href]").EachWithBreak(func (i int, sel *goquery.Selection) bool {
         var val string
         var ok bool
//         fmt.Printf("\n\nhref 1\n\n\n")
         if val, ok = sel.Attr("href"); ok {
//            fmt.Printf("\n\nhref 2\n\n\n")
            if len(val) >= 11 && val[:11] == "javascript:" {
//               fmt.Printf("\n\nhas js href\n\n\n")
               moreOptions = append(moreOptions, Prop_Scripted)
//               fmt.Printf("\n\nhref 3\n\n\n")
               return false
            }
         }
//         fmt.Printf("\n\nhref 4\n\n\n")
         return true
      })
   }

   doc.Find("[src]").EachWithBreak(func (i int, sel *goquery.Selection) bool {
      var val string
      var ok bool
      if val, ok = sel.Attr("src"); ok {
         if hasProto.MatchString(val) {
            moreOptions = append(moreOptions,Prop_RemoteResources)
            hasRemoteRef = true
//            fmt.Printf("\n\n%d %d->%s\n\n\n", sel.Nodes[0].Type, sel.Nodes[0].DataAtom, val)
            return false
         }
      }
      return true
   })

   if !hasRemoteRef {
      doc.Find("link[href]").EachWithBreak(func (i int, sel *goquery.Selection) bool {
         var val string
         var ok bool
         if val, ok = sel.Attr("href"); ok {
            if hasProto.MatchString(val) {
               moreOptions = append(moreOptions,Prop_RemoteResources)
//               fmt.Printf("\n\n%d %d->%s\n\n\n", sel.Nodes[0].Type, sel.Nodes[0].DataAtom, val)
               return false
            }
         }
         return true
      })
   }

//   fmt.Printf("moreOptions: %#v\n\n",moreOptions)
   return moreOptions
}


func init() {
   ASCDigitDash = regexp.MustCompile(`[^a-zA-Z0-9\-_/\.]`)
   hasProto = regexp.MustCompile(`^(?:(?:https?:)|(?:wss?:)|(?:ftp:))?//`)
}
