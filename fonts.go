package invoice

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// current root path of package
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// http://www.fpdf.org/en/doc/setfont.htm
// Symbol and ZapfDingbats were omitted
var nativelySupportedFonts = map[string]struct{}{
	"Arial":     {},
	"Courier":   {},
	"Helvetica": {},
	"Times":     {},
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		// file may or may not exist. See err for details.
		log.Print(err)
		return false
	}
}

func (pdf *PDF) mustAddFont(name, fileName, style string) error {
	// Quick return if the font is supported natively
	if _, ok := nativelySupportedFonts[fileName]; ok {
		// pdf.SetFont(name, style, fileName)
		switch name {
		case "bodyFont":
			bodyFont = fileName
		case "titleFont":
			titleFont = fileName
		}
		return nil
	}

	fontsPath := basepath + "/fonts/"
	e1 := Exists(fontsPath + fileName + ".json")
	e2 := Exists(fontsPath + fileName + ".z")

	if !e1 || !e2 {
		return fmt.Errorf("font %q (%q) could not be found at %q", fileName, name, fontsPath)
	}

	pdf.AddFont(name, style, "./"+fileName+".json")
	return nil
}

func (pdf *PDF) loadFonts(config *Configuration) error {
	// https://github.com/jung-kurt/gofpdf/issues/1#issuecomment-84037415
	pdf.SetFontLocation(basepath + "/fonts")
	if err := pdf.mustAddFont(bodyFont, config.FontBody, ""); err != nil {
		return err
	}
	if err := pdf.mustAddFont(bodyFont, config.FontBodyBold, "B"); err != nil {
		return err
	}
	if err := pdf.mustAddFont(titleFont, config.FontTitle, ""); err != nil {
		return err
	}

	return nil
}
