package epub30_test

func Example() {
   var err error
   var b ugarit.Book
   var target io.WriteCloser
   var gen ugarit.IndexGenerator
   var coverjpg *os.File
   var newId string

   newId = time.Now().Format("20060102150405")

   // The cover thumbnail is cached by iBooks and apparently, is referenced by
   // the epub filename. So, changing the filename when generating a new eBook
   // version (with a new cover) was the only way I found to make iBooks update
   // the book cover
   target, err = os.OpenFile("test-" + newId + ".epub", os.O_CREATE | os.O_RDWR, 0600)
   if err != nil {
      fmt.Printf("Open error: %s\n",err)
      os.Exit(0)
   }

   b, err = epub30.New(
      target,                            // io.Writer where to save the epub contents
      []string{"My Amazing Book Title"}, // eBook title
      []string{"en"},                    // eBook language
      []string{newId},                   // eBook ID
      []epub30.Author{
         epub30.Author{Data: "Me"},
         epub30.Author{Data: "Myself"},
      },                                 // eBook author(s)
      []string{"MY Dear Publisher"},     // eBook Publisher
      []epub30.Date{epub30.Date{Data:"2013-12-20"}},        // Date published
//      epub30.Signature{Href: "../META-INF/signatures.xml#AsYouLikeItSignature"},
      epub30.Signature{},
      []epub30.Metatag{ // Any metatags here
         epub30.Metatag{Property:"dcterms:dateCopyrighted", Data:"9999-01-01"},
      },
      "ltr", // PageProgression use "ltr" (left-to-right) or "rtl" (right-to-left)
      epub30.Versioner{}) // provide a versioner interface or use the package provided one
                          // which just generate a version string using the current time


   // Start the book first providing the cover
   coverjpg, err = os.Open("cover.jpg")
   if err != nil {
      fmt.Printf("Open jpeg error: %s\n",err)
      os.Exit(0)
   }

   _, _, err = b.AddCover(
      "cover.jpg",  // internal pathname where it will be stored
      "image/jpeg", // the mimetype
      coverjpg,     // the contents
      nil)
   if err != nil {
      fmt.Printf("AddCover error: %s\n",err)
      os.Exit(0)
   }


   // Then add pages to your book
   _, _, _, err = b.AddPage(
      "p1.xhtml",              // internal pathname where it will be stored
      "application/xhtml+xml", // the mimetype
      // the contents:
      // if provided it will be entirely read and closed
      // if not provided a new empty file will be created and an io.Writer will be returned
      strings.NewReader("<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>hello</body></html>"),
      "", // The optional page id, it will be generated if left empty
      &epub30.EPubOptions{
         TOCTitle: "Summary",           // TOC Title must be provided only once
         TOCItemTitle: "Hello Chapter", // Title for the page being added
      })
   if err != nil {
      fmt.Printf("AddPage error: %s\n",err)
      os.Exit(0)
   }



   // Create a TOC compatible with epub3
   gen, err = epub30.NewIndexGenerator()
   if err != nil {
      fmt.Printf("IndexGenerator 3.0 error: %s\n",err)
      os.Exit(0)
   }
   b.AddTOC(gen,"")

   // Create a TOC compatible with epub2
   gen, err = epub20.NewIndexGenerator(
      "en",  // Language
      newId, // Id
      "My Amazing Book Title", // Book title
      "Me", // Author
      b)    // The book object
   if err != nil {
      fmt.Printf("IndexGenerator 2.01 error: %s\n",err)
      os.Exit(0)
   }
   b.AddTOC(gen,"")


   // You MUST close the eBook, some important operations are done upon closing
   b.Close()

    // Output: Hello
}