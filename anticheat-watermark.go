package watermark

import (
	"fmt"
	"os"
	"sync"

	"github.com/unidoc/unidoc/pdf/creator"
	"github.com/unidoc/unidoc/pdf/model"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

type TextMeta struct {
	Font     string
	Fontsize float64
	X, Y     float64
	Color    string
	Angle    float64
}

func getFontByName(fontName string) fonts.Font {
	switch fontName {
	case "courier":
		return fonts.NewFontCourier()
	case "courier_bold":
		return fonts.NewFontCourierBold()
	case "courier_oblique":
		return fonts.NewFontCourierOblique()
	case "courier_bold_oblique":
		return fonts.NewFontCourierBoldOblique()
	case "helvetica":
		return fonts.NewFontHelvetica()
	case "helvetica_bold":
		return fonts.NewFontHelveticaBold()
	case "helvetica_oblique":
		return fonts.NewFontHelveticaOblique()
	case "helvetica_bold_oblique":
		return fonts.NewFontHelveticaBoldOblique()
	case "times":
		return fonts.NewFontTimesRoman()
	case "times_bold":
		return fonts.NewFontTimesBold()
	case "times_italic":
		return fonts.NewFontTimesItalic()
	case "times_bold_italic":
		return fonts.NewFontTimesBoldItalic()
	}

	return fonts.NewFontHelveticaBold()
}

func drawText(p *creator.Paragraph, c *creator.Creator, meta TextMeta) {
	p.SetWidth(p.Width())
	p.SetTextAlignment(creator.TextAlignmentCenter)
	p.SetFont(getFontByName(meta.Font))
	p.SetFontSize(meta.Fontsize)
	p.SetPos(meta.X, meta.Y)
	p.SetColor(creator.ColorRGBFromHex("#" + meta.Color))
	p.SetAngle(meta.Angle)

	_ = c.Draw(p)
}

func drawOnePage(c *creator.Creator, word string) {
	para := creator.NewParagraph(word)
	para.SetEnableWrap(false)

	for x := 0.0; x < c.Context().PageWidth; x += (para.Width() + 20) {
		for y := 0.0; y < c.Context().PageHeight; y += (para.Height() + 20) {
			meta := TextMeta{
				Font:     "courier",
				Fontsize: 12,
				X:        x,
				Y:        y,
				Color:    "E1E0E0",
				Angle:    0,
			}
			//			fmt.Println("Draw at", x, y)
			drawText(para, c, meta)
		}
	}
}

func DrawPDF(wg *sync.WaitGroup, pdffile string, word string, output string) {
	fmt.Println("Starting on: ", word)
	defer wg.Done()

	f, err := os.Open(pdffile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		panic(err)
	}
	c := creator.New()
	numPages, _ := pdfReader.GetNumPages()
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			panic(err)
		}

		block, err := creator.NewBlockFromPage(page)
		if err != nil {
			panic(err)
		}

		block.SetPos(0, 0)
		c.SetPageSize(creator.PageSize{block.Width(), block.Height()})

		c.NewPage()
		drawOnePage(c, word)
		if err = c.Draw(block); err != nil {
			panic(err)
		}
	}
	_ = c.WriteToFile(output)
	fmt.Println(word, " finished")
}

func DrawPDFSingle(pdffile string, word string, output string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go DrawPDF(&wg, pdffile, word, output)
	wg.Wait()
}
