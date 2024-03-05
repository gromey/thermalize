package thermalize

const (
	Left = iota
	Center
	Right
)

const (
	NoUnderling = iota
	OneDotUnderling
	TwoDotsUnderling
)

const (
	HRIFontA = iota // (12 x 24)
	HRIFontB        // (9 x 17)
)

const (
	HRINotPrinted = iota
	HRIAbove
	HRIBelow
	HRIAboveAndBelow
)

const (
	UpcA = iota
	UpcE
	JanEAN8
	JanEAN13
	Code39
	Code93
	Code128
	ITF
	NW7
	GS1128
	GS1Omnidirectional
	GS1Truncated
	GS1Limited
	GS1Expanded
)

const (
	L = iota // L recovers 7% of data
	M        // M recovers 15% of data
	Q        // Q recovers 25% of data
	H        // H recovers 30% of data
)

const (
	DrawerPin2 = iota
	DrawerPin5
)
