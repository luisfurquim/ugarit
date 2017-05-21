package epub30

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/luisfurquim/ugarit"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
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

	b.index = make([]*TOCContent, 0, 4)
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
					ndx:        pos,
					Title:      opt.TOCTitle,
					index:      make([]*TOCContent, 1, 4),
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

				tc.index = append(tc.index, &TOCContent{
					ndx:        pos,
					Title:      opt.TOCItemTitle,
					index:      make([]*TOCContent, 0, 4),
					subSection: ugarit.NewArabicNumbering("Chapter", true),
				})

				fmt.Printf("TOC: %#v\n", b.index)
			} else {
				tc = &TOCContent{
					ndx:        pos,
					Title:      opt.TOCItemTitle,
					index:      make([]*TOCContent, 0, 4),
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
			}
		default:
			return "", nil, nil, ugarit.ErrorInvalidOptionType

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
		src = strings.NewReader(fmt.Sprintf(svgcover, svgcontent))
		fileprop = prop[prop_CoverImage]

	} else if mimetype[:6] == "image/" {
		id, w, err = b.addFile(path, mimetype, src, "cover-image", opt, prop[prop_CoverImage])
		if err != nil {
			return "", w, err
		}

		mimetype = "application/xhtml+xml"
		src = strings.NewReader(fmt.Sprintf(imgcover, path))

		extension = filepath.Ext(path)
		pathhtml = path[:len(path)-len(extension)] + ".xhtml"

	} else {
		pathhtml = path
	}

	id, w, err = b.addFile(pathhtml, mimetype, src, "cover", opt, fileprop)
	if err != nil {
		return "", w, err
	}

	b.Package.Spine.Itemref = append(b.Package.Spine.Itemref, SpineItem{IDref: id, Linear: "yes"})
	b.Package.Guide.Reference = []Reference{Reference{Href: pathhtml, Type: "cover", Title: "Cover"}}

	return id, w, nil
}

// AddTOC saves the TOC file to the E-Book. Use it just before closing the E-Book.
// AddTOC creates/truncates the TOC file when the book is closed.
// If IndexGenerator != nil,  the generated html nodes are be passed through it before saved in the file.
// If id!="" it sets the TOC id.
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
					Val: fmt.Sprintf("pg%d", i+1),
				},
			},
		})
	}

	r, err = gen.GetDocument()
	if err != nil {
		//      fmt.Printf("\nError getting TOC: %s\n\n",err)
		return "", err
	}

	id, _, err = b.addFile(gen.GetPathName(), gen.GetMimeType(), r, id, &EPubOptions{}, gen.GetPropertyValue())
	fmt.Printf("\nSaved TOC: %s\n\n", err)

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
func (tc *TOCContent) ItemRef() string {
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
// AddFile creates/truncates the file specified by path,
// register as having the provided mimetype.
// If src is nil, it returns a valid io.Writer.
// If src is not nil, it copies its contents to the file and return a nil io.Writer.
// If path collides with a reserved path, it returns an error.
func (b *Book) AddFile(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, error) {
	var opt *EPubOptions
	var optProp string

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

	if opt != nil && opt.Prop > prop_Min && opt.Prop < prop_Max {
		optProp = prop[opt.Prop]
	}

	return b.addFile(path, mimetype, src, id, opt, optProp)
}

func (b *Book) addFile(path string, mimetype string, src io.Reader, id string, opt *EPubOptions, optProp string) (string, io.Writer, error) {
	//   if b.Package.Manifest == nil {
	//      b.Package.Manifest = []Manifest{}
	//   }

	if id == "" {
		id = fmt.Sprintf("pg%d", b.fid)
		b.fid++
	}

	b.Package.Manifest = append(b.Package.Manifest, Manifest{
		ID:         id,
		Href:       path,
		MediaType:  mimetype,
		Properties: optProp,
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

	err = b.zfd.Close()
	if err != nil {
		return err
	}

	return b.fd.Close()
}

// Creates an index generator object
func NewIndexGenerator() (*IndexGenerator, error) {
	var ig IndexGenerator
	var err error
	var node *html.Node

	ig.doc, err = goquery.NewDocumentFromReader(strings.NewReader(nav))
	if err != nil {
		return nil, err
	}

	ig.curr = ig.doc.Find("#TOClevel0").Nodes[0]

	node = &html.Node{
		Type: html.TextNode,
		Data: "Cover",
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
				Val: "cover.html",
			},
			html.Attribute{
				Key: "epub:type",
				Val: "cover",
			},
		},
	}

	ig.doc.Find("#lmarks").Nodes[0].AppendChild(
		&html.Node{
			FirstChild: node,
			LastChild:  node,
			Type:       html.ElementNode,
			DataAtom:   atom.Li,
			Data:       "li",
		})
	return &ig, nil
}

//        <li><a href="titlepg.xhtml" epub:type="titlepage">Title Page</a></li>
//        <li><a href="chapter.xhtml" epub:type="bodymatter">Start</a></li>
//        <li><a href="bibliog.xhtml" epub:type="bibliography">Bibliography</a></li>

// AddItem adds a new TOC entry
func (gen IndexGenerator) AddItem(item *html.Node) error {
	gen.curr.AppendChild(
		&html.Node{
			FirstChild: item,
			LastChild:  item,
			Type:       html.ElementNode,
			DataAtom:   atom.Li,
			Data:       "li",
		})
	return nil
}

// GetDocument returns the XHTML content of the TOC file
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

// GetPropertyValue returns the the value to set in the property attribute
// of the TOC OPF entry.
func (gen IndexGenerator) GetPropertyValue() string {
	return "nav"
}

// GetId returns the ID to use in the TOC OPF entry and the Spine TOC reference.
func (gen IndexGenerator) GetId() string {
	return "nav"
}

// Simple E-Book versioner
func (v Versioner) Next() string {
	return time.Now().Format("20060102150405")
}
