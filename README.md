Resize.cgi
==========

Image resizing tool which can be statically compiled to cgi/fcgi.


Deploy on a server running Apache httpd
=======================================

Example `.htaccess` file:

```
RewriteEngine on
RewriteCond %{QUERY_STRING} ^[0-9x]+$
RewriteCond %{REQUEST_FILENAME} !resize.cgi
RewriteRule !^resized/.* - [C]
RewriteRule ^(.*)$ resized/%{QUERY_STRING}/$1 [C]
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d [OR]
RewriteRule ^resized/[0-9x]+/(.*)$ resize.cgi/$1 [L,QSA]
```

How to build it?
================

You will need [Go](https://golang.org/) compiler.

You server is probably running Linux, so you need to prepare a binary for this platform:

```
GOOS=linux go build -tags cgi
```

How to use it?
==============

To automatically resize images, just add query string to image URL.

I.e.:

`image.jpg?800x600` will scale your image to 800 pixels width and 600 pixels height.

`image.jpg?800x` will scale your image to 800 pixels width and keep image proportions.

`image.jpg?x600` will scale your image to 600 pixels height and keep image proportions.
