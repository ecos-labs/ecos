package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/manifoldco/promptui"
)

// ===== Terminal Colors =====
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	CheckHeavy  = "✔"

	ClearPreviousLine = "\033[1A\033[K"
)

var terminalDetected *bool

// ===== Internal UI State =====

var isVerbose bool

// SetVerbose sets the global verbose mode for debug output.
func SetVerbose(verbose bool) {
	isVerbose = verbose
}

// ===== Helper Functions =====

func printWithSymbol(color, symbol, msg string) {
	fmt.Printf("%s%s%s %s\n", color, symbol, ColorReset, msg)
}

func printHeader(title, color string) {
	// Use RuneCountInString for proper Unicode character counting
	visualLength := max(utf8.RuneCountInString(title), 15)
	fmt.Printf("\n%s%s%s\n%s%s%s\n", color, title, ColorReset, ColorDim, strings.Repeat("─", visualLength), ColorReset)
}

// ===== Print Functions =====

// Print prints a formatted message to stdout.
func Print(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// PrintSuccess prints a success message with a checkmark symbol.
func PrintSuccess(msg string) {
	printWithSymbol(ColorGreen, CheckHeavy, msg)
}

// PrintError prints an error message with an X symbol.
func PrintError(msg string) {
	printWithSymbol(ColorRed, "✗", msg)
}

// PrintWarning prints a warning message with an exclamation symbol.
func PrintWarning(msg string) {
	printWithSymbol(ColorYellow, "!", msg)
}

// PrintInfo prints an informational message with an info symbol.
func PrintInfo(msg string) {
	printWithSymbol(ColorBlue, "i", msg)
}

// PrintDebug prints a debug message if verbose mode is enabled.
func PrintDebug(msg string) {
	if isVerbose {
		// Clear the current line to avoid conflicts with spinners
		fmt.Print("\r\033[K")
		printWithSymbol(ColorDim, "[>]", msg)
	}
}

// PrintProgress prints a progress message with a bullet symbol.
func PrintProgress(msg string) {
	printWithSymbol(ColorYellow, "[•]", msg+"...")
}

// PrintDryRun prints a dry-run message.
func PrintDryRun(msg string) {
	printWithSymbol(ColorCyan, "[DRY RUN]", msg)
}

// PrintStep prints a step message with step number and total.
func PrintStep(step, total int, msg string) {
	fmt.Printf("%s(%d/%d)%s %s\n", ColorCyan, step, total, ColorReset, msg)
}

// PrintHeader prints a header with bold styling.
func PrintHeader(title string) {
	printHeader(title, ColorBold)
}

// PrintSubHeader prints a subheader with purple styling.
func PrintSubHeader(title string) {
	printHeader(title, ColorPurple)
}

func PrintSectionHeader(title string) {
	fmt.Printf("\n%s%s%s\n", ColorCyan, title, ColorReset)
}

// supportsTrueColor checks if the terminal supports 24-bit true color RGB
func supportsTrueColor() bool {
	if !IsTerminal() {
		return false
	}

	// Check COLORTERM environment variable (explicit true color support)
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return true
	}

	// Check TERM for known terminals that support true color
	term := os.Getenv("TERM")
	trueColorTerms := []string{
		"xterm-kitty", "alacritty", "kitty", "wezterm", "vscode", "hyper",
		"iterm2", "iterm", // iTerm2 supports true color
	}
	for _, t := range trueColorTerms {
		if strings.Contains(term, t) {
			return true
		}
	}

	// macOS Terminal.app doesn't reliably support true color
	// iTerm2 is already checked above, so if we reach here on macOS,
	// it's likely Terminal.app which doesn't support true color well
	if runtime.GOOS == "darwin" {
		return false
	}

	// Windows Terminal supports true color
	if runtime.GOOS == "windows" {
		if os.Getenv("WT_SESSION") != "" {
			return true
		}
		if os.Getenv("ConEmuANSI") == "ON" {
			return true
		}
	}

	// Linux terminals with 256color often support true color
	if strings.Contains(term, "256color") {
		return true
	}

	return false
}

// getBannerLines returns the banner lines structure for the "ecos" ASCII art.
// This function is extracted for testability to ensure the exact spacing and structure is preserved.
func getBannerLines(eColor, cColor, oColor, sColor, reset string) [][]string {
	//nolint:dupl // Intentional duplication - banner lines are duplicated in test file for validation
	return [][]string{
		{"  ", eColor, "  ██████", reset, "   ", cColor, "  ██████", reset, "   ", oColor, " ███████", reset, "   ", sColor, "  ██████", reset},
		{"  ", eColor, " ██", reset, "  ", eColor, "  ██", reset, "  ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "  ", sColor, " ██", reset},
		{"  ", eColor, " ██", reset, "   ", eColor, " ██", reset, "  ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "  ", sColor, " ██", reset},
		{"  ", eColor, " ███████", reset, "   ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "  ", sColor, "  █████", reset},
		{"  ", eColor, " ██", reset, "        ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "       ", sColor, " ██", reset},
		{"  ", eColor, " ██", reset, "        ", cColor, " ██", reset, "       ", oColor, " ██", reset, "    ", oColor, " ██", reset, "       ", sColor, " ██", reset},
		{"  ", eColor, "  ██████", reset, "   ", cColor, "  ██████", reset, "   ", oColor, " ███████", reset, "   ", sColor, "  █████", reset},
	}
}

// PrintEcosBanner prints a large "ecos" ASCII art banner with gradient colors.
// It automatically detects terminal capabilities and uses true color if available,
// falling back to standard ANSI colors or plain text for non-terminal output.
func PrintEcosBanner() {
	useColors := IsTerminal()
	useTrueColor := useColors && supportsTrueColor()

	// Pre-compute color codes
	var eColor, cColor, oColor, sColor, reset string
	if useTrueColor {
		eColor = "\033[38;2;5;150;105m\033[1m"
		cColor = "\033[38;2;13;148;136m\033[1m"
		oColor = "\033[38;2;14;165;233m\033[1m"
		sColor = "\033[38;2;59;130;246m\033[1m"
		reset = "\033[0m"
	} else if useColors {
		eColor = ColorGreen + ColorBold
		cColor = ColorCyan + ColorBold
		oColor = ColorCyan + ColorBold
		sColor = ColorBlue + ColorBold
		reset = ColorReset
	}

	// Build banner and calculate width
	var bannerBuilder strings.Builder
	bannerLines := getBannerLines(eColor, cColor, oColor, sColor, reset)

	bannerBuilder.WriteByte('\n')
	bannerWidth := 0
	colorCodes := make(map[string]bool, 5)
	colorCodes[eColor] = true
	colorCodes[cColor] = true
	colorCodes[oColor] = true
	colorCodes[sColor] = true
	colorCodes[reset] = true
	for _, parts := range bannerLines {
		lineWidth := 0
		for _, part := range parts {
			if !colorCodes[part] {
				lineWidth += utf8.RuneCountInString(part)
			}
			bannerBuilder.WriteString(part)
		}
		if lineWidth > bannerWidth {
			bannerWidth = lineWidth
		}
		bannerBuilder.WriteByte('\n')
	}
	bannerBuilder.WriteByte('\n')
	fmt.Print(bannerBuilder.String())

	// Tagline and underline
	tagline := "Open FinOps Data Stack"
	taglineWidth := utf8.RuneCountInString(tagline)
	padding := (bannerWidth - taglineWidth) / 2
	if padding < 0 {
		padding = 0
	}
	pad := strings.Repeat(" ", padding)

	// Tagline
	if useColors {
		colors := []string{ColorGreen, ColorCyan, ColorCyan, ColorBlue}
		if useTrueColor {
			colors = []string{"\033[38;2;5;150;105m", "\033[38;2;13;148;136m", "\033[38;2;14;165;233m", "\033[38;2;59;130;246m"}
		}
		words := strings.Fields(tagline)
		var output strings.Builder
		output.WriteString(pad)
		for i, word := range words {
			if i > 0 {
				output.WriteString(" ")
			}
			output.WriteString(colors[i%len(colors)])
			output.WriteString(word)
			output.WriteString(reset)
		}
		fmt.Println(output.String())
	} else {
		fmt.Println(pad + tagline)
	}

	// Underline
	//nolint:gocritic // ifElseChain - if-else chain is clearer than switch for boolean conditions
	if useTrueColor {
		gradient := []string{"\033[38;2;5;150;105m", "\033[38;2;9;149;120m", "\033[38;2;13;148;136m", "\033[38;2;13;156;184m", "\033[38;2;14;165;233m", "\033[38;2;36;147;239m", "\033[38;2;59;130;246m"}
		var output strings.Builder
		output.WriteString(pad)
		for i := 0; i < taglineWidth; i++ {
			idx := (i * len(gradient)) / taglineWidth
			if idx >= len(gradient) {
				idx = len(gradient) - 1
			}
			output.WriteString(gradient[idx])
			output.WriteString("─")
		}
		output.WriteString(reset)
		fmt.Println(output.String())
	} else if useColors {
		fmt.Println(pad + ColorCyan + strings.Repeat("─", taglineWidth) + reset)
	} else {
		fmt.Println(pad + strings.Repeat("─", taglineWidth))
	}
}

// PrintFixedSelection prints a fixed selection with a checkmark symbol.
func PrintFixedSelection(title, value string) {
	fmt.Printf("%s✔%s %s: %s\n", ColorGreen, ColorReset, title, value)
}

// PrintVerbose prints a verbose message if verbose mode is enabled.
func PrintVerbose(msg string, verbose bool) {
	if verbose {
		printWithSymbol(ColorDim, "[v]", msg)
	}
}

// ===== Table and JSON Functions =====

func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = utf8.RuneCountInString(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellWidth := utf8.RuneCountInString(cell)
				if cellWidth > colWidths[i] {
					colWidths[i] = cellWidth
				}
			}
		}
	}

	// Print header
	fmt.Print(ColorBlue)
	for i, header := range headers {
		fmt.Printf("%-*s", colWidths[i]+2, header)
	}
	fmt.Println(ColorReset)

	// Separator
	fmt.Print(ColorBlue)
	for _, w := range colWidths {
		fmt.Print(strings.Repeat("-", w+2))
	}
	fmt.Println(ColorReset)

	// Rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf("%-*s", colWidths[i]+2, cell)
			}
		}
		fmt.Println()
	}
}

func PrintJSON(data any) {
	fmt.Printf("%s%v%s\n", ColorCyan, data, ColorReset)
}

// ===== Prompt Functions =====

func Select(label string, items []string, defaultIndex int, showTitle, indent bool) (int, string, error) {
	displayLabel := label
	if indent {
		displayLabel = "  " + label
	}

	selectPrompt := &promptui.Select{
		Label:     displayLabel,
		Items:     items,
		CursorPos: defaultIndex,
		Size:      len(items),
		HideHelp:  true,
	}

	idx, selectedValue, err := selectPrompt.Run()
	if err != nil {
		return idx, selectedValue, err
	}

	// Show title with checkmark if requested
	if showTitle {
		fmt.Print(ClearPreviousLine)
		prefix := ""
		if indent {
			prefix = "  "
		}
		// Clean up the selected value by removing extra spaces between words
		cleanValue := strings.Join(strings.Fields(selectedValue), " ")
		fmt.Printf("%s✔%s %s%s: %s\n", ColorGreen, ColorReset, prefix, label, cleanValue)
	}

	return idx, selectedValue, nil
}

func Input(label, defaultValue string, showTitle, indent bool, validate func(string) error) (string, error) {
	displayLabel := label
	if indent {
		displayLabel = "  " + label
	}

	prompt := promptui.Prompt{
		Label:   fmt.Sprintf("%s%s%s", ColorCyan, displayLabel, ColorReset),
		Default: defaultValue,
	}
	if validate != nil {
		prompt.Validate = validate
	}

	result, err := prompt.Run()
	if err != nil {
		return result, err
	}
	if result == "" {
		result = defaultValue
	}

	// Show title with checkmark if requested
	if showTitle {
		if IsTerminal() {
			fmt.Print("\r\033[K")        // Clear current line from cursor to end
			fmt.Print(ClearPreviousLine) // Clear previous line (in case promptui moved to new line)
		}
		prefix := ""
		if indent {
			prefix = "  "
		}
		fmt.Printf("%s✔%s %s%s: %s\n", ColorGreen, ColorReset, prefix, label, result)
	}

	return result, nil
}

func ConfirmPrompt(msg string) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s%s%s", ColorCyan, msg, ColorReset),
		IsConfirm: true,
		Default:   "N",
	}
	resp, err := prompt.Run()
	if err != nil {
		return false
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	return resp == "y" || resp == "yes"
}

func YesNo(label string, defaultNo bool) bool {
	options := []string{"Yes", "No"}
	defaultIdx := 1
	if !defaultNo {
		defaultIdx = 0
	}
	i, _, err := Select(label, options, defaultIdx, false, false)
	return err == nil && i == 0
}

// ===== Spinner =====

type Spinner struct {
	message string
	frames  []string
	delay   time.Duration
	active  bool
	done    chan struct{}
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay:   100 * time.Millisecond,
		done:    make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	if !IsTerminal() {
		fmt.Printf("[•] %s...\n", s.message)
		return
	}
	s.active = true
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				if !s.active {
					return
				}
				fmt.Printf("\r%s%s%s %s...", ColorCyan, s.frames[i%len(s.frames)], ColorReset, s.message)
				i++
				time.Sleep(s.delay)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	if !IsTerminal() || !s.active {
		return
	}
	s.active = false
	close(s.done)
	time.Sleep(s.delay)
	fmt.Print("\r\033[K")
}

func (s *Spinner) Success(msg string) {
	s.Stop()
	if msg == "" {
		msg = s.message
	}
	PrintSuccess(msg)
}

func (s *Spinner) Error(msg string) {
	s.Stop()
	if msg == "" {
		msg = s.message + " failed"
	}
	PrintError(msg)
}

// ===== ProgressBar =====

type ProgressBar struct {
	weights        []int
	totalWeight    int
	currentWeight  int
	completedSteps int
	totalSteps     int
	width          int
	message        string
	complete       bool
	autoComplete   bool
}

func NewProgressBar(total int, message string) *ProgressBar {
	weights := make([]int, total)
	for i := range weights {
		weights[i] = 1
	}
	return NewWeightedProgressBar(weights, message)
}

func NewWeightedProgressBar(weights []int, message string) *ProgressBar {
	total := 0
	for _, w := range weights {
		total += w
	}
	pb := &ProgressBar{
		weights:      weights,
		totalWeight:  total,
		totalSteps:   len(weights),
		width:        40,
		message:      message,
		autoComplete: true,
	}

	// Set up automatic completion when all steps are done
	if pb.autoComplete {
		// This will be called when the ProgressBar goes out of scope
		// or when the function returns
		runtime.SetFinalizer(pb, (*ProgressBar).autoFinalize)
	}

	return pb
}

// EnableAutoComplete enables automatic completion when all steps are done
func (p *ProgressBar) EnableAutoComplete() *ProgressBar {
	p.autoComplete = true
	return p
}

// DisableAutoComplete disables automatic completion - requires manual Complete() call
func (p *ProgressBar) DisableAutoComplete() *ProgressBar {
	p.autoComplete = false
	runtime.SetFinalizer(p, nil) // Remove finalizer
	return p
}

// Finish should be called with defer to ensure proper completion
func (p *ProgressBar) Finish() {
	if !p.complete {
		if p.completedSteps >= p.totalSteps || p.currentWeight >= p.totalWeight {
			p.Complete()
		} else {
			p.Render()
		}
	}
}

// autoFinalize is called by the runtime finalizer
func (p *ProgressBar) autoFinalize() {
	if !p.complete && p.autoComplete {
		p.Finish()
	}
}

// AdvanceStep increments progress by the weight of the step
func (p *ProgressBar) AdvanceStep(step int, stepMessage string) {
	if step < len(p.weights) {
		p.currentWeight += p.weights[step]
		p.completedSteps++
	}
	if stepMessage != "" {
		printWithSymbol(ColorGreen, CheckHeavy, stepMessage)
	}

	// Auto-complete if all steps are done and auto-complete is enabled
	if p.autoComplete && p.completedSteps >= p.totalSteps {
		p.Complete()
	}
}

// Complete renders the final progress bar at 100%.
func (p *ProgressBar) Complete() {
	if p.complete {
		return // Already completed
	}

	if p.currentWeight < p.totalWeight {
		p.currentWeight = p.totalWeight
	}
	p.complete = true
	p.render()
	fmt.Println() // Only print newline at final completion

	// Remove finalizer since we're done
	runtime.SetFinalizer(p, nil)
}

func (p *ProgressBar) render() {
	percent := float64(p.currentWeight) / float64(p.totalWeight)
	filled := int(percent * float64(p.width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)
	fmt.Printf("\r[%s] %3.0f%% %s", bar, percent*100, p.message)
}

// Render shows the progress bar at current progress without forcing completion
func (p *ProgressBar) Render() {
	if p.complete {
		return // Already completed
	}
	p.render()
	fmt.Println()
}

// ===== Logger =====

type Logger struct {
	prefix  string
	verbose bool
}

func NewLogger(prefix string, verbose bool) *Logger {
	return &Logger{prefix: prefix, verbose: verbose}
}

func (l *Logger) logWithColor(color string, format string, args ...any) {
	fmt.Printf("%s%s%s %s\n", color, l.prefix, ColorReset, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(format string, args ...any) { l.logWithColor(ColorBlue, format, args...) }

func (l *Logger) Success(format string, args ...any) {
	l.logWithColor(ColorGreen, format, args...)
}

func (l *Logger) Warning(format string, args ...any) {
	l.logWithColor(ColorYellow, format, args...)
}
func (l *Logger) Error(format string, args ...any) { l.logWithColor(ColorRed, format, args...) }
func (l *Logger) Debug(format string, args ...any) {
	if l.verbose {
		l.logWithColor(ColorWhite, format, args...)
	}
}

// IsTerminal checks if stdout is connected to a terminal.
// Results are cached after the first check for performance.
func IsTerminal() bool {
	if terminalDetected != nil {
		return *terminalDetected
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		val := false
		terminalDetected = &val
		return val
	}
	val := (info.Mode() & os.ModeCharDevice) != 0
	terminalDetected = &val
	return val
}
