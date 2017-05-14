package epub20


var imgcover string = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <title>Cover</title>
  <style type="text/css">
img {
   padding: 0;
   margin: 0;
   height: 95%%;
   height: 95vh;
   max-width: 100%%;
}
  </style>
 </head>
 <body style="margin: 0; padding: 0;">
  <div id="cover" style="text-align: center;display: block;height: 95%%;">
   <img src="%s"/>
  </div>
 </body>
</html>`

// dtb:uid = b.Package.Metadata.Identifier[0]
var ncx string = `<?xml version="1.0" encoding="UTF-8" ?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx version="2005-1" xml:lang="en" xmlns="http://www.daisy.org/z3986/2005/ncx/">
   <head><meta name="dtb:uid" content=" /></head>

   <docTitle>
      <text>Metamorphosis </text>
   </docTitle>

   <docAuthor>
      <text>Franz Kafka</text>
   </docAuthor>

   <navMap>
      <navPoint id="front-cover" playOrder="1">
         <navLabel><text>Cover</text></navLabel>
         <content src="OEBPS/front-cover.html" />
      </navPoint>
      <navPoint id="title-page" playOrder="2">
         <navLabel><text>Title Page</text></navLabel>
         <content src="OEBPS/title-page.html" />
      </navPoint>
   </navMap>
</ncx>`



var svgcover string = `<?xml version="1.0" ?>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <style type="text/css">
section {
   text-align: center;
   display: block;
   height: 95%%;
}

body {
   margin: 0;
   padding: 0;
}
  </style>
 </head>
 <body>
  <section id="cover" epub:type="cover">%s</section>
 </body>
</html>`


var rootfolder string = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">` +
 `<rootfiles>` +
  `<rootfile full-path="%s/content.opf" media-type="application/oebps-package+xml" />` +
 `</rootfiles>` +
`</container>`
