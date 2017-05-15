package epub30

var imgcover string = `<?xml version="1.0" ?>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
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

var nav string = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE html>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <title class="title"></title>
 </head>
 <body>
  <nav epub:type="toc" id="toc"><ol id="TOClevel0"></ol></nav>
  <nav epub:type="landmarks"><ol id="lmarks"></ol></nav>
 </body>
</html>`

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
