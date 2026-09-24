package internal

// Page formats and orientations the config accepts.
const (
	PageFormatA4         = "a4"
	OrientationPortrait  = "portrait"
	OrientationLandscape = "landscape"
)

// pageSides returns the sheet's width and height as CSS lengths. Anything but
// landscape is portrait, so a zero-value Config still gets a valid sheet.
func pageSides(orientation string) (width, height string) {
	if orientation == OrientationLandscape {
		return "297mm", "210mm"
	}
	return "210mm", "297mm"
}

// legacyOrientation maps the retired document_max_width to an orientation:
// only the A4 long side meant landscape; every other width meant portrait.
func legacyOrientation(maxWidth string) string {
	if maxWidth == "297mm" {
		return OrientationLandscape
	}
	return OrientationPortrait
}
