package epub20

import (
   "io"
   "fmt"
   "strings"
   "io/ioutil"
   "archive/zip"
   "encoding/xml"
   "path/filepath"
   "golang.org/x/net/html"
   "golang.org/x/net/html/atom"
   "github.com/luisfurquim/ugarit"
)


// Create a blank EPub Book
func New(
      target            io.WriteCloser,
      Title           []string,
      Language        []string,
      identifier      []string,
      Creator         []Author,
      Publisher       []string,
      Date            []Date,
      signature         Signature,
      metatag         []Metatag,
      PageProgression   string) (*Book, error) {

   var b Book
   var Sig *Signature
   var lang string
   var ids []Identifier

   b.RootFolder = "OEBPS"

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
   _, err = f.Write([]byte(fmt.Sprintf(rootfolder,b.RootFolder)))
   if err != nil {
       return nil, err
   }


   if (len(metatag)>0) {
      if (len(Language)>0) {
         lang = strings.Join(Language,",")
      } else {
         lang = "en"
      }

      for i, _ := range metatag {
         metatag[i].Langattr = lang
      }
   }


   ids = make([]Identifier,len(identifier))
   for i, id := range identifier {
      ids[i] = Identifier{
         Data: id,
         ID: "pub-id" ,
      }
   }

   b.Package = Package{
      Version: "2.0",
      Xmlns:   "http://www.idpf.org/2007/opf",
      UID:     "pub-id",
      Metadata: Metadata{
         Xmlns:      "http://purl.org/dc/elements/1.1/",
         Title:      Title,
         Language:   Language,
         Identifier: ids,
         Creator:    Creator,
         Publisher:  Publisher,
         Date:       Date,
         Signature:  Sig,
         Metatag:    metatag,
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

   b.subSection = ugarit.NewArabicNumbering("Chapter",true)

   return &b, nil
}

// AddPage saves the page to the E-Book and adds an entry in the TOC pointing to it
// It calls AddFile. So, if you use this method, you don't need to call AddFile to save it
// otherwise it will be stored twice in the E-Book file.
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

//   fmt.Printf("OPTIONS: %#v\n",options)

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


// AddCover saves the Cover in the E-Book. Use it before saving any content to the E-Book
func (b *Book) AddCover(path string, mimetype string, src io.Reader, options interface{}) (string, io.Writer, error) {
   var w io.Writer
   var err error
   var opt *EPubOptions
   var id string
   var svgcontent []byte
   var pathhtml string
   var extension string
   var fileprop string

  if options != nil {
      switch options.(type) {
         case *EPubOptions:
            opt = options.(*EPubOptions)
         default:
            return "", nil, ugarit.ErrorInvalidOptionType
      }
   }

   b.Package.Metadata.Metatag = append(b.Package.Metadata.Metatag,Metatag{
      Name:"cover",
      Content:"cover-image",
   })

   if mimetype == "image/svg+xml" {
      svgcontent, err = ioutil.ReadAll(src)
      if err != nil {
         return "", w, err
      }

      extension = filepath.Ext(path)
      pathhtml = path[:len(path)-len(extension)] + ".xhtml"
      src = strings.NewReader(fmt.Sprintf(svgcover,svgcontent))

   } else if mimetype[:6] == "image/" {
      id, w, err = b.addFile(path, mimetype, src, "cover-image", opt, "")
      if err != nil {
         return "", w, err
      }

      mimetype = "application/xhtml+xml"
      src = strings.NewReader(fmt.Sprintf(imgcover,path))

      extension = filepath.Ext(path)
      pathhtml = path[:len(path)-len(extension)] + ".xhtml"

   } else {
      pathhtml = path
   }

   id, w, err = b.addFile(pathhtml, mimetype, src, "cover", opt, fileprop)
   if err != nil {
      return "", w, err
   }

   b.Package.Spine.Itemref = append(b.Package.Spine.Itemref,SpineItem{IDref: id, Linear:"no"})
   b.Package.Guide.Reference = []Reference{Reference{Href: pathhtml, Type:"cover", Title:"Cover"}}

   return id, w, nil
}

// AddTOC saves the TOC file to the E-Book. Use it just before closing the E-Book
func (b *Book) AddTOC(gen ugarit.IndexGenerator, id string) (string, error) {
   var err error
   var r io.Reader
   var txt *html.Node

   if id == "" {
      id = gen.GetId()
   }

   for i, ndx := range b.index {
      txt = &html.Node{
         Type: html.TextNode,
         Data: ndx.Title,
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
               Val: b.Package.Manifest[ndx.ndx].Href,
            },
            html.Attribute{
               Key: "id",
               Val: fmt.Sprintf("pg%d",i+1),
            },
         },
      })
   }

   r, err = gen.GetDocument()
   if err != nil {
//      fmt.Printf("\nError getting TOC: %s\n\n",err)
      return "", err
   }

   id, _, err = b.addFile(gen.GetPathName(), gen.GetMimeType(), r, "ncx", nil, "")

   return id, err
}

// TOCChild retrieves the Nth child of the TOC top level
func (b *Book) TOCChild(n int) ugarit.TOC {
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
func (tc *TOCContent) TOCChild(n int) ugarit.TOC {
   if n < len(tc.index) {
      return tc.index[n]
   }

   return nil
}

// TOCLen retrieves the item count in this TOC section
func (tc *TOCContent) TOCLen() int {
   return len(tc.index)
}

// ItemRef retrieves the reference ID of the TOC Item
func (tc *TOCContent) ItemRef() string{
   return tc.ref
}

// ItemTitle retrieves the item title (label) as it appears in the TOC
func (tc *TOCContent) ItemTitle() string {
   return tc.Title
}

// ContentRef retrieves the path o the file pointed by the TOC entry
func (tc *TOCContent) ContentRef() string {
   return tc.manifest[tc.ndx].Href
}

// SubSectionStyle sets the object to style this TOC subsection
func (tc *TOCContent) SubSectionStyle(sty ugarit.SectionStyle) {
   tc.subSection = sty
}







// AddFile stores the file in the E-Book. It does not include it in the TOC.
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

    if id == "cover"{
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

   return b.addFile(path, mimetype, src, id, opt, optProp)
}

func (b *Book) addFile(path string, mimetype string, src io.Reader, id string, opt *EPubOptions, optProp string) (string, io.Writer, error) {
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
      io.Copy(w,src)
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


   err = b.zfd.Close()
   if err != nil {
      return err
   }

   return b.fd.Close()
}

// Creates an index generator object
func NewIndexGenerator(lang string, uid string, tit string, auth string, b ugarit.Book) (*IndexGenerator, error) {
   var ig IndexGenerator

   b.SpineAttr("toc","ncx")

   ig.doc = Ncx{
      Version: "2005-1",
      Langattr: lang,
      Xmlns: "http://www.daisy.org/z3986/2005/ncx/",
      Metatag: []Metatag{
         Metatag{
            Name: "dtb:uid",
            Content: uid,
         },
         Metatag{
            Name: "cover",
            Content: "cover",
         },
      },
      Title: tit,
      Author: auth,
      Points:    []NavPoint{},
//      Pages:     []PageTarget{},
   }

   ig.curr = &ig.doc.Points
   return &ig, nil
}

// AddItem adds a new TOC entry
func (gen *IndexGenerator) AddItem(item *html.Node) error {
   *gen.curr = append(*gen.curr,NavPoint{
      Id:        item.Attr[1].Val,
      PlayOrder: item.Attr[1].Val[2:],
      Label:     item.FirstChild.Data,
      Content:   Content{
         Src: item.Attr[0].Val,
      },
      Points:  []NavPoint{},
   })
   return nil
}

// GetDocument returns the XHTML content of the TOC file
func (gen *IndexGenerator) GetDocument() (io.Reader, error) {
   var doc []byte
   var err error

   doc, err = xml.Marshal(gen.doc)
   if err != nil {
      return nil, err
   }

   return strings.NewReader(xml.Header + string(doc)), nil
}



// GetMimeType returns the mimetype of the TOC file.
func (gen *IndexGenerator) GetMimeType() string {
   return "application/x-dtbncx+xml"
}

// GetPathName returns the relative pathname of the TOC file
func (gen *IndexGenerator) GetPathName() string {
   return "toc.ncx"
}

// GetPropertyValue returns the the value to set in the property attribute
// of the TOC OPF entry. It is not defined in the EPub 2.x specs.
func (gen IndexGenerator) GetPropertyValue() string {
   return ""
}

// GetId returns the ID to use in the TOC OPF entry and the Spine TOC reference.
func (gen IndexGenerator) GetId() string {
   return "ncx"
}


