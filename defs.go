package ugarit

import (
   "io"
   "fmt"
   "errors"
   "strings"
   "golang.org/x/net/html"
   "github.com/StefanSchroeder/Golang-Roman"
)

type Book interface {
   // All pathnames must be considered absolute if path[0]=='/', relative otherwise

   // AddPage must create/truncate the file specified by path,
   // register as having the provided mimetype and making to figure in the book index.
   // If src is nil, it has to return a valid io.Writer.
   // If src is not nil, it must copy its content to the file and return a nil io.Writer.
   // If path collides with a reserved path, it must return an error.
   // If the page is to be added to the TOC, it must return a TOCRef object pointing
   // to the TOC entry that refers to the added page, otherwise, TOCRef must be nil
   AddPage(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, TOCRef, error)

   // AddFile must create/truncate the file specified by path,
   // register as having the provided mimetype.
   // If src is nil, it has to return a valid io.WriteCloser.
   // If src is not nil, it must copy its contents to the file and return a nil io.Writer.
   // If path collides with a reserved path, it must return an error.
   AddFile(path string, mimetype string, src io.Reader, id string, options interface{}) (string, io.Writer, error)

   // AddTOC must create/truncate the TOC file.
   // The file generated must be in html format.
   // If IndexGenerator != nil,  the generated html nodes
   // must be passed through it before saved in the file.
   // If id!="" it must set the TOC id
   AddTOC(gen IndexGenerator, id string) (string, error)

   // AddCover must create/truncate the file specified by path,
   // register as having the provided mimetype and making it be the book cover.
   // If src is nil, it has to return a valid io.Writer.
   // If src is not nil, it must copy its contents to the file and return a nil io.Writer.
   // If path collides with a reserved path, it must return an error.
   AddCover(path string, mimetype string, src io.Reader, options interface{}) (string, io.Writer, error)

   // SpineAttr must set any spine attribute named as key with val
   // Error checking is allowed but it must fail silently.
   //  This method may be deprecated if someday it is considered
   // too much epub addicted
   SpineAttr(key, val string)

   // Closes the Ebook
   Close() error

/*
   //Remove removes path
   Remove(path string) error

   // RemoveAll removes path and any children it contains. It removes everything it can
   // but returns the first error it encounters. If the path does not exist, RemoveAll
   // returns nil (no error).
   RemoveAll(path string) error

   //Rename renames (moves) oldpath to newpath
   Rename(oldpath, newpath string) error

   // Opens path...
   Open(path string) (io.reader, error)

   // List all files (absolute pathnames) in the ebook
   AllFiles() []string

*/
}

type IndexGenerator interface {
   // AddItem must append the item to the end of the table of contents.
   // The item is provided as a link (the HTML 'A' element) with the
   // href attribute pointing to the page file, the text content of the
   // element has the title of the page and, if provided when AddPage
   // called, the attribute id.
   AddItem(item *html.Node) error

   // GetDocument must return the entire TOC document
   GetDocument() (io.Reader, error)

   // GetMimeType returns the mimetype of the TOC file.
   GetMimeType() string

   // GetPathName returns the relative pathname of the TOC file
   GetPathName() string

   // GetPropertyValue returns the property value for the TOC
   // E.G. EPub 3.0 returns 'nav' and EPub 2.o returns ''
   GetPropertyValue() string

   // GetId returns the Id value for the TOC file in the manifest
   // E.G. EPub 3.0 returns 'nav' and EPub 2.o returns 'ncx'
   GetId() string
}

type TOC interface {
   TOCChild(n int) TOC
   TOCLen() int
}

type TOCItem interface {
   ItemRef() string
   ItemTitle() string
   ContentRef() string
   SubSectionStyle(SectionStyle)
}

type TOCRef interface {
   TOC
   TOCItem
}

type SectionStyle interface {
   Prefix() string
   Number(root string, number int) string
}

type Versioner interface {
   // The versioner compliant object musst provide a Next method
   // which must provide a new unique version string each time
   // it is called
   Next() string
}

type ArabicNumbering struct {
   pfx string
   hierarchy bool
}

type UpperRomanNumbering struct {
   pfx string
   hierarchy bool
}

type LowerRomanNumbering struct {
   pfx string
   hierarchy bool
}

type UpperLetterNumbering struct {
   pfx string
   hierarchy bool
}

type LowerLetterNumbering struct {
   pfx string
   hierarchy bool
}


func (ss ArabicNumbering) Prefix() string {
   return ss.pfx
}

func (ss UpperRomanNumbering) Prefix() string {
   return ss.pfx
}

func (ss LowerRomanNumbering) Prefix() string {
   return ss.pfx
}

func (ss UpperLetterNumbering) Prefix() string {
   return ss.pfx
}

func (ss LowerLetterNumbering) Prefix() string {
   return ss.pfx
}

func (ss ArabicNumbering) Number(root string, number int) string {
   if ss.hierarchy {
      return fmt.Sprintf("%s.%d",root,number)
   }
   return fmt.Sprintf("%d",number)
}

func (ss UpperRomanNumbering) Number(root string, number int) string {
   if ss.hierarchy {
      return fmt.Sprintf("%s.%s",root,roman.Roman(number))
   }
   return fmt.Sprintf("%s",roman.Roman(number))
}


func (ss LowerRomanNumbering) Number(root string, number int) string {
   if ss.hierarchy {
      return fmt.Sprintf("%s.%s",root,strings.ToLower(roman.Roman(number)))
   }
   return fmt.Sprintf("%s",strings.ToLower(roman.Roman(number)))
}


func (ss UpperLetterNumbering) Number(root string, number int) string {
   if ss.hierarchy {
      return fmt.Sprintf("%s.%c",root,'A' + number)
   }
   return fmt.Sprintf("%c",'A' + number)
}


func (ss LowerLetterNumbering) Number(root string, number int) string {
   if ss.hierarchy {
      return fmt.Sprintf("%s.%c",root,'a' + number)
   }
   return fmt.Sprintf("%c",'a' + number)
}

func NewArabicNumbering(prefix string, useHierarchy bool) ArabicNumbering {
   return ArabicNumbering{pfx: prefix, hierarchy: useHierarchy}
}

func NewUpperRomanNumbering(prefix string, useHierarchy bool) UpperRomanNumbering {
   return UpperRomanNumbering{pfx: prefix, hierarchy: useHierarchy}
}

func NewLowerRomanNumbering(prefix string, useHierarchy bool) LowerRomanNumbering {
   return LowerRomanNumbering{pfx: prefix, hierarchy: useHierarchy}
}

func NewUpperLetterNumbering(prefix string, useHierarchy bool) UpperLetterNumbering {
   return UpperLetterNumbering{pfx: prefix, hierarchy: useHierarchy}
}

func NewLowerLetterNumbering(prefix string, useHierarchy bool) LowerLetterNumbering {
   return LowerLetterNumbering{pfx: prefix, hierarchy: useHierarchy}
}

var ErrorInvalidOptionType error = errors.New("Invalid option type")
var ErrorInvalidPathname error = errors.New("Invalid pathname")
var ErrorTOCItemTitleNotFound error = errors.New("TOC item title not found")
var ErrorReservedId error = errors.New("Reserved Id")

