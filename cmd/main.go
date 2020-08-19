package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	watermark "github.com/hare1039/anticheat-watermark"
	quote "github.com/kballard/go-shellquote"
)

func main() {
	var pdffile, namefile string
	if len(os.Args) != 3 {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Please enter the pdf file: ")
		scanner.Scan()

		pdflist, err := quote.Split(strings.TrimSpace(scanner.Text()))
		if err != nil {
			panic(err)
		}
		pdffile = pdflist[0]

		fmt.Print("Please enter the name file: ")
		scanner.Scan()
		namefilelist, err := quote.Split(strings.TrimSpace(scanner.Text()))
		if err != nil {
			panic(err)
		}
		namefile = namefilelist[0]
	} else {
		pdffile = os.Args[1]
		namefile = os.Args[2]
	}

	if strings.Contains(pdffile, "pdf") {
	} else if strings.Contains(namefile, "pdf") {
		pdffile, namefile = namefile, pdffile
	} else {
		fmt.Println("Can't find 'pdf' in argv")
		fmt.Scanln()
		return
	}

	fp, err := os.Open(namefile)
	if err != nil {
		fmt.Println(err, "Abort.")
		fmt.Scanln()
		panic(err)
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)

	fullName := strings.Split(filepath.Base(pdffile), ".")[0]
	os.Chdir(filepath.Dir(pdffile))

	err = os.Mkdir(fullName, 0755)
	if err != nil {
		fmt.Println("ERROR: folder ", fullName, " exist. Abort.")
		fmt.Scanln()
		panic(err)
	}

	fmt.Println("Start adding anti-cheat watermark")
	var wg sync.WaitGroup
	for scanner.Scan() {
		name := scanner.Text()
		wg.Add(1)
		go watermark.DrawPDF(&wg, pdffile, name, fullName+"/"+name+".pdf")
	}
	wg.Wait()
}
