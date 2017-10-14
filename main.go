package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/nfnt/resize"
)

var query = regexp.MustCompile(`(\d*)x(\d*)`)

func handler(w http.ResponseWriter, r *http.Request) error {
	var width, heigth uint
	params := query.FindStringSubmatch(r.URL.RawQuery)
	if len(params) == 0 || params[0] == "x" {
		e := fmt.Sprintf("width or height not found in URL parameters '%s'", r.URL.RawQuery)
		http.Error(w, e, http.StatusBadRequest)
		return nil
	}
	if params[1] != "" {
		fmt.Sscan(params[1], &width)
	}
	if params[2] != "" {
		fmt.Sscan(params[2], &heigth)
	}
	inFilename := os.Getenv("PATH_INFO")
	if inFilename == "" {
		inFilename = r.URL.Path
	}
	if inFilename[0] == '/' {
		inFilename = inFilename[1:]
	}
	outFilename := path.Join("resized", params[0], inFilename)
	if file, err := os.Open(outFilename); err == nil {
		defer file.Close()
		w.Header().Set("X-Resize-Cache", "HIT")
		return serveFile(w, r, file)
	}
	in, err := os.Open(inFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return nil
	}
	defer in.Close()
	img, format, err := image.Decode(in)
	if err != nil {
		return err
	}
	imgSize := img.Bounds().Max
	if (width != 0 && width < uint(imgSize.X)) || (heigth != 0 && heigth < uint(imgSize.Y)) {
		img = resize.Resize(width, heigth, img, resize.Lanczos3)
	}
	if err := os.MkdirAll(path.Dir(outFilename), 0750); err != nil {
		return err
	}
	out, err := os.Create(outFilename + "~")
	if err != nil {
		return err
	}
	defer out.Close()
	switch format {
	case "jpeg":
		err = jpeg.Encode(out, img, nil)
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(out, img)
	case "gif":
		err = gif.Encode(out, img, nil)
	}
	if err != nil {
		return err
	}
	out.Sync()
	inStat, err := in.Stat()
	if err != nil {
		return err
	}
	outStat, err := out.Stat()
	if err != nil {
		return err
	}
	w.Header().Set("X-Resize-Cache", "MISS")
	if inStat.Size() < outStat.Size() {
		rel, err := filepath.Rel(path.Dir(outFilename), inFilename)
		if err != nil {
			return err
		}
		if err := os.Symlink(rel, outFilename); err != nil {
			return err
		}
		return serveFile(w, r, in)
	}
	if err := os.Rename(outFilename+"~", outFilename); err != nil {
		return err
	}
	return serveFile(w, r, out)
}

func serveFile(w http.ResponseWriter, r *http.Request, file *os.File) error {
	etag := Etag(file)
	if etag != "" {
		w.Header().Set("ETag", etag)
		if etag == r.Header.Get("If-None-Match") {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
	}
	contentType := mime.TypeByExtension(path.Ext(file.Name()))
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.Copy(w, file); err != nil {
		return err
	}
	return nil
}

func Etag(file *os.File) string {
	stat, err := file.Stat()
	if err != nil {
		return ""
	}
	return fmt.Sprintf(`W/"%x-%x"`, stat.Size(), stat.ModTime().UnixNano()/1000)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	http.Handle("/", Handler(handler))
	log.Fatalln(Serve(http.DefaultServeMux))
}

type Handler func(http.ResponseWriter, *http.Request) error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err == nil {
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
