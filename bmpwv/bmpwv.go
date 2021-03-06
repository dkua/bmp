// bmpwv.go (c) 2013 David Rook

package main

import (
	// Alien imports
	"github.com/hotei/bmp"
	"github.com/russross/blackfriday" // Russ Ross 2012-11-22 version
	// go 1.X only below here
	//"bufio"
	"bytes"
	"fmt"
	"image/png"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	//"time"
)

const (
	hostIPstr  = "127.0.0.1"
	portNum    = 8282
	bmpURL     = "/bmp/"
	mdURL      = "/md/"
	serverRoot = "/www/"
)

var (
	portNumString = fmt.Sprintf(":%d", portNum)
	listenOnPort  = hostIPstr + portNumString
	g_fileNames   []string
)

var myBmpDir = []byte{}
var myMdDir = []byte{}

func checkBmpName(pathname string, info os.FileInfo, err error) error {
	//fmt.Printf("checking %s\n", pathname)
	if info == nil {
		fmt.Printf("WARNING --->  no stat info: %s\n", pathname)
		os.Exit(1)
	}
	if info.IsDir() {
		// return filepath.SkipDir
	} else { // regular file
		//fmt.Printf("found %s %s\n", pathname, filepath.Ext(pathname))
		if filepath.Ext(pathname) == ".bmp" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
		}
	}
	return nil
}

func checkMdName(pathname string, info os.FileInfo, err error) error {
	fmt.Printf("checking %s\n", pathname)
	if info == nil {
		fmt.Printf("WARNING --->  no stat info: %s\n", pathname)
		os.Exit(1)
	}
	if info.IsDir() {
		// return filepath.SkipDir
		// g_fileNames = append(g_fileNames, pathname)
		return nil
	} else { // regular file
		//fmt.Printf("found %s %s\n", pathname, filepath.Ext(pathname))
		ext := filepath.Ext(pathname)
		if ext == ".md" || ext == ".markdown" || ext == ".mdown" {
			//fmt.Printf("appending\n")
			g_fileNames = append(g_fileNames, pathname)
		}
	}
	return nil
}

// show thumbnail as clickable link to original image
func makeBmpLine(s string) []byte {
	//return []byte(fmt.Sprintf("<a href=\"%s\">View %s</a><br>\n",s,s))
	workDir := serverRoot + bmpURL[1:]
	s = s[len(workDir):]
	return []byte(fmt.Sprintf("<a href= \"%s\"><img src=\"%s\" height=100 width=150 align=center> %s</a><br>\n", s, s, s))
}

func makeMdLine(s string) []byte {
	workDir := serverRoot + mdURL[1:]
	s = s[len(workDir):]
	return []byte(fmt.Sprintf("<a href=\"%s\">%s</a><br>", s, s))
}

func init() {
	pathName := serverRoot + bmpURL[1:]
	g_fileNames = make([]string, 0, 20)
	myBmpDir = []byte(`<html><!-- comment --><head><title>Test BMP package</title></head><body>click on image to see in original size<br>`) // {}
	stats, err := os.Stat(pathName)
	if err != nil {
		fmt.Printf("Can't get fileinfo for %s\n", pathName)
		os.Exit(1)
	}
	if stats.IsDir() {
		filepath.Walk(pathName, checkBmpName)
	} else {
		fmt.Printf("this argument must be a directory (but %s isn't)\n", pathName)
		os.Exit(-1)
	}
	//fmt.Printf("g_fileNames = %v\n", g_fileNames)
	for _, val := range g_fileNames {
		//fmt.Printf("%v\n", val)
		line := makeBmpLine(val)
		myBmpDir = append(myBmpDir, line...)
	}
	t := []byte(`</body></html>`)
	myBmpDir = append(myBmpDir, t...)

	pathName = serverRoot + mdURL[1:]
	g_fileNames = make([]string, 0, 20)
	myMdDir = []byte(`<html><!-- comment --><head><title>Test MD package</title></head><body>click to read<br>`) // {}
	stats, err = os.Stat(pathName)
	if err != nil {
		fmt.Printf("Can't get fileinfo for %s\n", pathName)
		os.Exit(1)
	}
	if stats.IsDir() {
		filepath.Walk(pathName, checkMdName)
	} else {
		fmt.Printf("this argument must be a directory (but %s isn't)\n", pathName)
		os.Exit(-1)
	}
	//fmt.Printf("g_fileNames = %v\n", g_fileNames)
	for _, val := range g_fileNames {
		//fmt.Printf("%v\n", val)
		line := makeMdLine(val)
		myMdDir = append(myMdDir, line...)
	}
	t = []byte(`</body></html>`)
	myMdDir = append(myMdDir, t...)
}

func mdHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == mdURL {
		w.Write(myMdDir)
		return
	}
	var output []byte
	var err error
	workDir := serverRoot + mdURL[1:]
	fileName := workDir + r.URL.Path[len(mdURL):]
	fmt.Printf("mdHandler: fname = %s\n", fileName)
	ext := filepath.Ext(fileName)
	if ext == ".md" || ext == ".markdown" || ext == ".mdown" {
		output = htmlFromMd(fileName)
	} else {
		output, err = ioutil.ReadFile(fileName)
		if err != nil {
			errStr := fmt.Sprintf("mdHandler: %v\n", err)
			fmt.Printf("%s\n", errStr)
			w.Write([]byte(fmt.Sprintf("404 - Not Found\n")))
			return
		}
	}
	w.Write(output)
}

func htmlFromMd(fname string) []byte {
	var output []byte
	input, err := ioutil.ReadFile(fname)
	if err != nil {
		tmp := fmt.Sprintf("Problem reading input, can't open %s", fname)
		output = []byte(tmp)
	} else {
		output = blackfriday.MarkdownCommon(input) // MarkdownBasic(input) also possible
	}
	if false { // debug use only
		os.Stdout.Write(input)
		os.Stdout.Write(output)
	}
	return output
}

// working and pretty
func bmpHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == bmpURL {
		w.Write(myBmpDir)
		return
	}
	workDir := serverRoot + bmpURL[1:]
	fmt.Printf("workDir(%s)\n", workDir)
	fmt.Printf("r.URL.Path(%s)\n", r.URL.Path)
	imageName := workDir + r.URL.Path[len(bmpURL):]
	fmt.Printf("bmpHandler: imageName = %s\n", imageName)
	bf, err := os.Open(imageName)
	if err != nil {
		fmt.Printf("bmpHandler: cant open bmp %s\n", imageName)
		return
	}
	img, err := bmp.Decode(bf)
	if err != nil {
		fmt.Printf("bmpHandler: bmp decode failed for %s\n", imageName)
		w.Write([]byte(fmt.Sprintf("Decode failed for %s\n", imageName)))
		return
	}
	b := make([]byte, 0, 10000)
	wo := bytes.NewBuffer(b)
	err = png.Encode(wo, img)
	if err != nil {
		fmt.Printf("bmpHandler: png encode failed for %s\n", imageName)
		return
	}
	w.Write(wo.Bytes())
}

// ---------------------------------------------------------------------------     M A I N

func main() {
	//	http.HandleFunc(virtualURL, html)
	http.HandleFunc(mdURL, mdHandler)
	http.HandleFunc(bmpURL, bmpHandler)
	// Handle(serverRoot, is like a dir missing an index "ftp-style"
	http.Handle(serverRoot, http.StripPrefix(serverRoot, http.FileServer(http.Dir(serverRoot))))
	log.Printf("bmpwv is ready to serve at %s\n", listenOn)
	err := http.ListenAndServe(listenOnPort, nil)
	if err != nil {
		log.Printf("bmpwv: error running webserver %v", err)
	}
}
