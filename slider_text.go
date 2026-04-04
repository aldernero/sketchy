package sketchy

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var intSliderLiteralRegexp = regexp.MustCompile(`^[+-]?[0-9]+$`)

func fractionDigitsForFloatText(sl *FloatSlider) int {
	if sl.TextDecimals >= 0 {
		return sl.TextDecimals
	}
	return sl.digits
}

func shouldUseScientificFloatDisplay(v float64) bool {
	if v == 0 {
		return false
	}
	av := math.Abs(v)
	return av < 1e-4 || av >= 1e6
}

func formatFloatSliderDisplay(sl *FloatSlider) string {
	v := sl.Val
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	if shouldUseScientificFloatDisplay(v) {
		return strconv.FormatFloat(v, 'g', -1, 64)
	}
	prec := fractionDigitsForFloatText(sl)
	return strconv.FormatFloat(v, 'f', prec, 64)
}

// parseValidFloatString accepts Go float literals including scientific notation (e.g. 1e-12).
func parseValidFloatString(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

func parseValidIntSliderString(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" || !intSliderLiteralRegexp.MatchString(s) {
		return 0, false
	}
	v64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	if v64 > int64(math.MaxInt) || v64 < int64(math.MinInt) {
		return 0, false
	}
	return int(v64), true
}

func snapIntSliderToStep(sl *IntSlider, v int) int {
	v = clampInt(v, sl.MinVal, sl.MaxVal)
	if sl.Incr <= 1 {
		return v
	}
	k := int(math.Round(float64(v-sl.MinVal) / float64(sl.Incr)))
	out := sl.MinVal + k*sl.Incr
	return clampInt(out, sl.MinVal, sl.MaxVal)
}

func formatIntSliderDisplay(sl *IntSlider) string {
	return strconv.Itoa(sl.Val)
}

func (sl *FloatSlider) syncTextBufFromVal() {
	sl.textBuf = formatFloatSliderDisplay(sl)
	sl.textSyncVal = sl.Val
	sl.textSyncOK = true
}

func (sl *FloatSlider) maybeSyncTextBufFromVal() {
	if !sl.textSyncOK || sl.textSyncVal != sl.Val {
		sl.syncTextBufFromVal()
	}
}

func commitFloatSliderText(sl *FloatSlider) {
	s := strings.TrimSpace(sl.textBuf)
	if s == "" {
		sl.syncTextBufFromVal()
		return
	}
	v, ok := parseValidFloatString(s)
	if !ok {
		sl.syncTextBufFromVal()
		return
	}
	sl.Val = clampFloat(v, sl.MinVal, sl.MaxVal)
	sl.syncTextBufFromVal()
}

func (sl *IntSlider) syncTextBufFromVal() {
	sl.textBuf = formatIntSliderDisplay(sl)
	sl.textSyncVal = sl.Val
	sl.textSyncOK = true
}

func (sl *IntSlider) maybeSyncTextBufFromVal() {
	if !sl.textSyncOK || sl.textSyncVal != sl.Val {
		sl.syncTextBufFromVal()
	}
}

func commitIntSliderText(sl *IntSlider) {
	s := strings.TrimSpace(sl.textBuf)
	if s == "" {
		sl.syncTextBufFromVal()
		return
	}
	v, ok := parseValidIntSliderString(s)
	if !ok {
		sl.syncTextBufFromVal()
		return
	}
	sl.Val = snapIntSliderToStep(sl, v)
	sl.syncTextBufFromVal()
}
