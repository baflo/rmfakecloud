package exporter

import (
	"bytes"
	"fmt"
	"io"

	rmc "github.com/joagonca/rmc-go"
	"github.com/unidoc/unipdf/v3/model"
)

// ExportV6ToPdfNative converts v6 .rm file to PDF using rmc-go library (in-process)
// This uses the Cairo renderer for native PDF generation
func ExportV6ToPdfNative(rmData []byte, output io.Writer) error {
	opts := &rmc.Options{
		UseLegacy: false, // Always use Cairo renderer (not Inkscape)
	}

	// Convert from bytes to PDF bytes
	pdfData, err := rmc.ConvertFromBytes(rmData, rmc.FormatPDF, opts)
	if err != nil {
		return fmt.Errorf("failed to convert v6 rm to PDF: %w", err)
	}

	// Write to output
	_, err = io.Copy(output, bytes.NewReader(pdfData))
	if err != nil {
		return fmt.Errorf("failed to write PDF output: %w", err)
	}

	return nil
}

// ExportV6ToSvgNative converts v6 .rm file to SVG using rmc-go library
func ExportV6ToSvgNative(rmData []byte, output io.Writer) error {
	opts := &rmc.Options{}

	svgData, err := rmc.ConvertFromBytes(rmData, rmc.FormatSVG, opts)
	if err != nil {
		return fmt.Errorf("failed to convert v6 rm to SVG: %w", err)
	}

	_, err = io.Copy(output, bytes.NewReader(svgData))
	if err != nil {
		return fmt.Errorf("failed to write SVG output: %w", err)
	}

	return nil
}

// ExportV6MultiPageToPdfNative converts multiple v6 .rm pages to a single PDF
// Each page is converted individually with rmc-go, then merged using unipdf
// This preserves the original page dimensions from each converted page
func ExportV6MultiPageToPdfNative(pages [][]byte, output io.Writer) error {
	if len(pages) == 0 {
		return fmt.Errorf("no pages provided")
	}

	// Single page optimization
	if len(pages) == 1 {
		return ExportV6ToPdfNative(pages[0], output)
	}

	// Convert all pages to individual PDFs
	var pdfReaders []*model.PdfReader
	for pageNum, pageData := range pages {
		// Convert this page to PDF
		var pdfBuf bytes.Buffer
		err := ExportV6ToPdfNative(pageData, &pdfBuf)
		if err != nil {
			return fmt.Errorf("failed to convert page %d: %w", pageNum, err)
		}

		// Load PDF for merging
		pdfReader, err := model.NewPdfReader(bytes.NewReader(pdfBuf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to read PDF for page %d: %w", pageNum, err)
		}

		pdfReaders = append(pdfReaders, pdfReader)
	}

	// Merge all PDFs into one
	pdfWriter := model.NewPdfWriter()

	for pageNum, pdfReader := range pdfReaders {
		numPages, err := pdfReader.GetNumPages()
		if err != nil {
			return fmt.Errorf("failed to get page count for page %d: %w", pageNum, err)
		}

		// Each converted page should be a single-page PDF
		for i := 0; i < numPages; i++ {
			page, err := pdfReader.GetPage(i + 1)
			if err != nil {
				return fmt.Errorf("failed to get page %d from PDF %d: %w", i+1, pageNum, err)
			}

			err = pdfWriter.AddPage(page)
			if err != nil {
				return fmt.Errorf("failed to add page %d to output: %w", pageNum, err)
			}
		}
	}

	// Write the merged PDF
	err := pdfWriter.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write merged PDF: %w", err)
	}

	return nil
}
