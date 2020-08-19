package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	watermark "github.com/hare1039/anticheat-watermark"
)

// copy from https://golangcode.com/create-zip-files-in-go/

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func Generate(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["files"]

	dir, err := ioutil.TempDir("", "anticheat")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Mkdir: ", dir)
	defer os.RemoveAll(dir)

	var pdffile, namefile string
	for _, file := range files {
		filename := dir + "/" + filepath.Base(file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		if strings.Contains(filename, "pdf") {
			pdffile = filename
		} else {
			namefile = filename
		}
	}

	fp, err := os.Open(namefile)
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()

	err = os.Mkdir(dir+"/generated", 0755)
	if err != nil {
		log.Println(err)
		return
	}

	var AllGeneratedPDF []string
	scanner := bufio.NewScanner(fp)
	var wg sync.WaitGroup
	for scanner.Scan() {
		name := scanner.Text()
		wg.Add(1)
		AllGeneratedPDF = append(AllGeneratedPDF, dir+"/generated/"+name+".pdf")
		go watermark.DrawPDF(&wg, pdffile, name, dir+"/generated/"+name+".pdf")
	}
	wg.Wait()

	fullName := strings.Split(filepath.Base(pdffile), ".")[0]
	ZipFiles("static/"+fullName+".zip", AllGeneratedPDF)
	c.HTML(http.StatusOK, "complete.tmpl", gin.H{
		"downloadURL": "/static/" + url.PathEscape(fullName+".zip"),
	})
}

func MainPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func main() {
	r := gin.Default()
	r.GET("/", MainPage)
	r.POST("/generate", Generate)
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	r.Run(":9000")
}
