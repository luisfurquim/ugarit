package epub30

var Imgcover string = `<?xml version="1.0" ?>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <title>Cover</title>
  <style type="text/css">
BODY {
	margin: 0;
	padding: 0;
}

#cover {
	text-align: center;
	display: block;
	height: 95%%;
   padding: 0;
   margin: 0;
}

#cover img {
   padding: 0;
   margin: 0;
   height: 95%%;
   height: 95vh;
   max-width: 100%%;
}
  </style>
 </head>
 <body><div id="cover"><img src="%s"/></div></body>
</html>`

var ImgcoverSized string = `<?xml version="1.0" ?>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <title>Cover</title>
  <meta name="viewport" content="width=%d, height=%d"></meta> 
  <style type="text/css">
BODY {
   padding: 0;
	width: %dpx;
	height: %dpx;
	margin-left: %dpx;
	margin-right: %dpx;
	margin-top: %dpx;
	margin-bottom: %dpx;
}

#cover {
	text-align: center;
	display: block;
	height: 95%%;
   padding: 0;
   margin: 0;
}

#cover img {
   padding: 0;
	width: %dpx;
	height: %dpx;
	margin-left: -%dpx;
	margin-right: -%dpx;
	margin-top: -%dpx;
	margin-bottom: -%dpx;
}
  </style>
 </head>
 <body><div id="cover"><img src="%s"/></div></body>
</html>`

var ImgcoverLessSized string = `<?xml version="1.0" ?>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <title>Cover</title>
  <meta name="viewport" content="width=%d, height=%d"></meta> 
  <style type="text/css">
BODY {
	width: %dpx;
	height: %dpx;
	margin-left: %dpx;
	margin-right: %dpx;
	margin-top: %dpx;
	margin-bottom: %dpx;
}

img {
   padding: 0;
   margin: 0;
   height: 95%%;
   height: 95vh;
   max-width: 100%%;
}
  </style>
 </head>
 <body><div id="cover" style="text-align: center;display: block;height: 95%%;"><img src="%s"/></div></body>
</html>`

/*

var nav string = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE html>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta content="text/html; charset=UTF-8"/>
  <title class="title"></title>
  <style type="text/css">
  DIV {
   margin-left: 3em;
   text-indent: 3em;
  }
  </style>
</head>
<body>
  <p><a id="toc"></a></p>
  <div class="title"></div>
  <div> <br/></div>
  <div id="TOClevel0"></div>
  <ol id="lmarks"></ol>
 </body>
</html>`

*/

var nav string = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE html>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <title class="title">%s</title>
  <style type="text/css">
  OL {
    margin-left: 3em;
    text-indent: 3em;
  }
  </style>
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

var svgcoverSized string = `<?xml version="1.0" ?>
<html xmlns:epub="http://www.idpf.org/2007/ops" xmlns="http://www.w3.org/1999/xhtml">
 <head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=%d, height=%d"></meta> 
  <style type="text/css">
section {
   text-align: center;
   display: block;
   height: 95%%;
}

BODY {
   padding: 0;
	width: %dpx;
	height: %dpx;
	margin-left: %dpx;
	margin-right: %dpx;
	margin-top: %dpx;
	margin-bottom: %dpx;
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

var iBooksFonts string = `<?xml version="1.0" encoding="UTF-8"?>
<display_options><platform name="*"><option name="specified-fonts">true</option></platform></display_options>`

