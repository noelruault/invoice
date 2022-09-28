package invoice

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/jung-kurt/gofpdf"
	yaml "gopkg.in/yaml.v3"
)

var (
	titleFont = "titleFont"
	bodyFont  = "bodyFont"

	//go:embed fonts/iso-8859-15.map
	fontDescriptorFileBytes []byte

	// current root path of package
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

type Configuration struct {
	DateFormat         string   `yaml:"dateFormat"`
	FontBody           string   `yaml:"fontBody"`
	FontBodyBold       string   `yaml:"fontBodyBold"`
	FontDescriptorFile string   `yaml:"fontDescriptorFile"`
	FontTitle          string   `yaml:"fontTitle"`
	Net                bool     `yaml:"net"`
	TextDate1          string   `yaml:"textDate1"`
	TextDate2          string   `yaml:"textDate2"`
	TextFooterTitle    string   `yaml:"textFooterTitle"`
	TextTableColumns   []string `yaml:"textTableColumns"`
	TextTitle          string   `yaml:"textTitle"`
}

func NewConfig(path string) (*Configuration, error) {
	c := Configuration{
		FontBody:     "Arial",
		FontBodyBold: "Arial",
		FontTitle:    "Arial",

		DateFormat: "02/01/2006",

		TextTitle:        "INVOICE",
		TextDate1:        "Issue date",
		TextDate2:        "Due date",
		TextFooterTitle:  "Payment details",
		TextTableColumns: []string{"Description", "Quantity", "Price", "Line Total"},
	}

	if path == "" {
		return &c, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Rune is used to unmarshal the currency string from the YAML file into a rune,
// which will be needed when printing the currency symbol
type Rune rune

func (r *Rune) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}

	rn, _ := utf8.DecodeRune([]byte(s))
	*r = Rune(rn)
	return nil
}

type Invoice struct {
	Account            string
	Client             string
	ConfigFile         string
	Currency           Rune
	Emitted, Delivered time.Time
	ID                 string
	Info               string
	Name               string
	Net                bool
	Services           []Service
	SumTax             bool
	Taxes              []Tax
}

func (i *Invoice) Total() float64 {
	var t float64

	for _, s := range i.Services {
		t += s.Amount()
	}

	return t
}

type Tax struct {
	Description string
	Value       float64
}

type Service struct {
	Description string
	UnitCost    float64
	Unit        string
	Quantity    float64
}

func formatWithSymbol(value float64, symbol Rune) string {
	return fmt.Sprintf("%s %c", strconv.FormatFloat(value, 'f', -1, 64), symbol)
}

func getTotalTaxes(taxes []Tax) float64 {
	var totalTax float64
	for _, tax := range taxes {
		totalTax += tax.Value
	}
	return totalTax
}

func (s *Service) format(currency Rune, taxes []Tax, net bool) (desc, qty, uc, amount string) {
	desc = s.Description
	qty = fmt.Sprintf("%g %s", s.Quantity, s.Unit)
	if net {
		ucValue := s.UnitCost - s.UnitCost*getTotalTaxes(taxes)/100
		uc = formatWithSymbol(ucValue, currency)
		amountValue := s.UnitCost*s.Quantity - s.UnitCost*s.Quantity*getTotalTaxes(taxes)/100
		amount = formatWithSymbol(amountValue, currency)
	} else {
		uc = formatWithSymbol(s.UnitCost, currency)
		amount = formatWithSymbol(s.UnitCost*s.Quantity, currency)
	}

	return
}

func (s *Service) Amount() float64 {
	return s.UnitCost * s.Quantity
}

type PDF struct {
	*gofpdf.Fpdf
}

func (i *Invoice) PDF(configFile, output string) (*string, error) {
	config, err := NewConfig(configFile)
	if err != nil {
		return nil, err
	}

	pdf := PDF{gofpdf.New("P", "mm", "A4", "")}
	if err := pdf.loadFonts(config); err != nil {
		return nil, err
	}

	pdf.AddPage()

	var tr func(string) string
	if config.FontDescriptorFile == "" {
		tr = pdf.UnicodeTranslatorFromDescriptor(config.FontDescriptorFile)
	} else {
		tr, err = gofpdf.UnicodeTranslator(bytes.NewReader(fontDescriptorFileBytes))
		if err != nil {
			return nil, err
		}
	}

	pdf.SetXY(20, 20)
	pdf.SetFont(bodyFont, "", 12)
	pdf.MultiCell(180, 5, tr(i.Client), "", "R", false)

	pdf.SetXY(20, 15)
	pdf.SetFont(bodyFont, "B", 14)
	pdf.Write(0, tr(fmt.Sprintf("%s", i.Name)))

	pdf.SetXY(20, 20)
	pdf.SetFont(bodyFont, "", 12)
	pdf.MultiCell(0, 5, tr(i.Info), "", "", false)
	pdf.Ln(50)

	pdf.SetXY(20, 80)
	pdf.SetFont(titleFont, "", 40)
	pdf.Write(0, tr(fmt.Sprintf("%s", config.TextTitle)))

	pdf.SetFont(bodyFont, "", 20)
	pdf.SetXY(20, 95)
	pdf.Write(0, tr(fmt.Sprintf("#%s", i.ID)))

	pdf.SetXY(20, 110)
	pdf.SetFontSize(10)

	var dateStr string
	dateStr += config.TextDate1 + ":  " + i.Emitted.Format(config.DateFormat) + "\n"
	dateStr += config.TextDate2 + ":  " + i.Delivered.Format(config.DateFormat) + "\n"
	pdf.MultiCell(0, 5, tr(dateStr), "", "", false)

	var lineNumber int
	// makeLine := func(description, quantity, unitCost, amount string) {
	makeLine := func(columns ...string) {
		if len(columns) > 4 {
			log.Fatal("number of columns for main table must be 4")
		}

		if lineNumber == 0 {
			pdf.SetFont(bodyFont, "B", 12)
		} else {
			pdf.SetFont(bodyFont, "", 12)
		}

		pdf.SetX(20)
		pdf.Cell(80, 5, tr(columns[0]))
		pdf.CellFormat(30, 5, tr(columns[1]), "", 0, "R", false, 0, "")
		pdf.CellFormat(30, 5, tr(columns[2]), "", 0, "R", false, 0, "")
		pdf.CellFormat(30, 5, tr(columns[3]), "", 0, "R", false, 0, "")

		if lineNumber == 0 {
			pdf.Ln(6)
			pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
			pdf.Ln(2)
		} else {
			pdf.Ln(6)
		}

		lineNumber++
	}

	pdf.Ln(10)
	makeLine(config.TextTableColumns...)

	for _, s := range i.Services {
		makeLine(s.format(i.Currency, i.Taxes, i.Net))
	}

	// Add tax info
	if len(i.Taxes) > 0 {
		pdf.Ln(10)
		pdf.Line(110, pdf.GetY(), 190, pdf.GetY())

		for _, tax := range i.Taxes { // convert to map string float
			pdf.Ln(5)
			pdf.SetXY(100, pdf.GetY())
			pdf.CellFormat(30, 5, tr(tax.Description), "", 0, "R", false, 0, "")
			// pdf.CellFormat(30, 5, tr("+"+formatPercentatge(tax.Value)), "", 0, "R", false, 0, "")
			pdf.CellFormat(30, 5, tr("+"+formatWithSymbol(tax.Value, '%')), "", 0, "R", false, 0, "")
			pdf.CellFormat(30, 5, tr("+"+formatWithSymbol((i.Total()*tax.Value/100), i.Currency)), "", 0, "R", false, 0, "")
		}
		pdf.Ln(10)
		pdf.Line(110, pdf.GetY(), 190, pdf.GetY())

	}

	pdf.Ln(10)
	pdf.SetFont(bodyFont, "B", 12)

	var totalAmount string
	if i.Net {
		totalAmount = formatWithSymbol(i.Total(), i.Currency)
	} else {
		totalAmount = formatWithSymbol(i.Total()+i.Total()*getTotalTaxes(i.Taxes)/100, i.Currency)
	}
	pdf.MultiCell(180, 5, tr("Total : "+totalAmount), "", "R", false)

	pdf.SetXY(20, 245)
	pdf.SetFont(bodyFont, "B", 12)
	pdf.Write(0, tr(fmt.Sprintf("%s", config.TextFooterTitle)))

	pdf.SetXY(20, 250)
	pdf.SetFont(bodyFont, "", 11)
	pdf.MultiCell(0, 5, tr(i.Account), "", "L", false)

	pdf.SetXY(180, 275)
	pdf.Write(0, tr("Page 1/1"))

	if output == "" {
		output = i.ID + ".pdf"
	}

	err = pdf.OutputFileAndClose(output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func Generate(config, input, output string) (*string, error) {
	data, err := ioutil.ReadFile(input)
	if err != nil {
		return nil, err
	}

	i := Invoice{
		ID:        time.Now().Format("20060102"),
		Emitted:   time.Now(),
		Delivered: time.Now().AddDate(0, 0, 5),
	}
	err = yaml.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	return i.PDF(config, output)
}
