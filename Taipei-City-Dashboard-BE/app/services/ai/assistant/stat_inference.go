package assistant

import (
	"fmt"
	"math"
	"sort"
)

func buildHypothesisCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	test := stringOption(input, "test", "welch")
	groups := groupPointsBySeries(points)
	if test == "chi_square_2x2" || test == "fisher_2x2" {
		return buildCategoricalCard(ds, groups, test)
	}
	if len(groups) != 2 {
		return AnalysisCard{}, fmt.Errorf("welch test requires exactly two series")
	}
	names, left, right := twoGroups(groups)
	if len(left) < 2 || len(right) < 2 {
		return AnalysisCard{}, fmt.Errorf("welch test requires at least 2 values per series")
	}
	result := welchResult(left, right)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, len(points))
	findings := []string{
		fmt.Sprintf("%s 平均 %.4g，%s 平均 %.4g。", names[0], result["mean_a"], names[1], result["mean_b"]),
		fmt.Sprintf("差異 %.4g，Welch t=%.4g，df=%.4g。", result["difference"], result["t_statistic"], result["df"]),
	}
	assumptions := append(baseAssumptions(), "p 值使用常態近似；正式報告需用統計套件交叉覆核。")
	result["visualization"] = effectIntervalVisualization(result)
	return statCard("hypothesis_test", "兩組 Welch 比較已完成", findings, assumptions, ds, confidence, result, "95% CI for mean difference"), nil
}

func buildCategoricalCard(ds statDataset, groups map[string][]statPoint, test string) (AnalysisCard, error) {
	table, names, err := contingency2x2(groups)
	if err != nil {
		return AnalysisCard{}, err
	}
	chi2, p := chiSquare2x2(table)
	fisherP := fisherExact2x2(table)
	result := map[string]interface{}{
		"test":               test,
		"series":             names,
		"table":              table,
		"chi_square":         round4(chi2),
		"chi_square_p_value": round4(p),
		"fisher_p_value":     round4(fisherP),
	}
	result["visualization"] = contingencyVisualization(table, names, fisherP)
	findings := []string{
		fmt.Sprintf("2x2 table for %s/%s completed。", names[0], names[1]),
		fmt.Sprintf("chi-square %.4g，Fisher exact p %.4g。", chi2, fisherP),
	}
	assumptions := append(baseAssumptions(), "類別檢定只在 2x2 非負計數資料上執行。")
	return statCard("hypothesis_test", "2x2 類別檢定已完成", findings, assumptions, ds, "medium", result, "not_applicable"), nil
}

func twoGroups(groups map[string][]statPoint) ([]string, []float64, []float64) {
	names := make([]string, 0, 2)
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, statValues(groups[names[0]]), statValues(groups[names[1]])
}

func welchResult(a, b []float64) map[string]interface{} {
	meanA, meanB := mean(a), mean(b)
	varA, varB := variance(a), variance(b)
	se := math.Sqrt(varA/float64(len(a)) + varB/float64(len(b)))
	diff := meanA - meanB
	t := 0.0
	if se > 0 {
		t = diff / se
	}
	dfNum := math.Pow(varA/float64(len(a))+varB/float64(len(b)), 2)
	dfDen := math.Pow(varA/float64(len(a)), 2)/float64(len(a)-1) +
		math.Pow(varB/float64(len(b)), 2)/float64(len(b)-1)
	df := 0.0
	if dfDen > 0 {
		df = dfNum / dfDen
	}
	return map[string]interface{}{
		"mean_a": meanA, "mean_b": meanB, "difference": diff,
		"standard_error": se, "t_statistic": t, "df": df,
		"p_value_normal_approx": math.Erfc(math.Abs(t) / math.Sqrt2),
		"ci_low":                diff - 1.96*se,
		"ci_high":               diff + 1.96*se,
	}
}

func contingency2x2(groups map[string][]statPoint) ([2][2]int, []string, error) {
	var table [2][2]int
	names := make([]string, 0, 2)
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) != 2 {
		return table, names, fmt.Errorf("2x2 categorical test requires exactly two series")
	}
	for i, name := range names {
		points := sortStatPoints(groups[name])
		if len(points) != 2 {
			return table, names, fmt.Errorf("2x2 categorical test requires two values per series")
		}
		for j, point := range points {
			if point.Value < 0 || math.Trunc(point.Value) != point.Value {
				return table, names, fmt.Errorf("2x2 categorical values must be non-negative integers")
			}
			table[i][j] = int(point.Value)
		}
	}
	return table, names, nil
}

func chiSquare2x2(t [2][2]int) (float64, float64) {
	a, b, c, d := float64(t[0][0]), float64(t[0][1]), float64(t[1][0]), float64(t[1][1])
	n := a + b + c + d
	denom := (a + b) * (c + d) * (a + c) * (b + d)
	if denom == 0 {
		return 0, 1
	}
	chi2 := n * math.Pow(a*d-b*c, 2) / denom
	return chi2, math.Erfc(math.Sqrt(chi2 / 2))
}

func fisherExact2x2(t [2][2]int) float64 {
	row1, row2 := t[0][0]+t[0][1], t[1][0]+t[1][1]
	col1 := t[0][0] + t[1][0]
	observed := hypergeometricProb(t[0][0], row1, row2, col1)
	minA := maxInt(0, col1-row2)
	maxA := minInt(row1, col1)
	p := 0.0
	for a := minA; a <= maxA; a++ {
		prob := hypergeometricProb(a, row1, row2, col1)
		if prob <= observed+1e-12 {
			p += prob
		}
	}
	return math.Min(1, p)
}

func hypergeometricProb(a, row1, row2, col1 int) float64 {
	c := col1 - a
	return math.Exp(logChoose(row1, a) + logChoose(row2, c) - logChoose(row1+row2, col1))
}

func logChoose(n, k int) float64 {
	if k < 0 || k > n {
		return math.Inf(-1)
	}
	lnN, _ := math.Lgamma(float64(n + 1))
	lnK, _ := math.Lgamma(float64(k + 1))
	lnNK, _ := math.Lgamma(float64(n - k + 1))
	return lnN - lnK - lnNK
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
