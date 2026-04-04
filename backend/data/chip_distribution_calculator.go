package data

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type ChipDistributionResult struct {
	LatestDate   string
	LatestClose  float64
	SampleDays   int
	BenefitPart  float64
	AvgCost      float64
	Cost70       DistributionInterval
	Cost90       DistributionInterval
	Distribution []DistributionBucket
}

type DistributionInterval struct {
	LowerPrice    float64
	UpperPrice    float64
	Concentration float64
}

type DistributionBucket struct {
	Price  float64
	Weight float64
}

type parsedKLine struct {
	day      string
	open     float64
	close    float64
	high     float64
	low      float64
	turnover float64
}

func CalculateChipDistribution(klines []KLineData, accuracyFactor int) (*ChipDistributionResult, error) {
	if len(klines) == 0 {
		return nil, fmt.Errorf("K 线数据不能为空")
	}
	if accuracyFactor <= 0 {
		return nil, fmt.Errorf("accuracyFactor 必须大于 0")
	}

	parsed := make([]parsedKLine, len(klines))
	minPrice := math.MaxFloat64
	maxPrice := 0.0
	for i, k := range klines {
		open, err := parseFloatField(k.Open, "Open")
		if err != nil {
			return nil, err
		}
		close, err := parseFloatField(k.Close, "Close")
		if err != nil {
			return nil, err
		}
		high, err := parseFloatField(k.High, "High")
		if err != nil {
			return nil, err
		}
		low, err := parseFloatField(k.Low, "Low")
		if err != nil {
			return nil, err
		}
		if high < low {
			return nil, fmt.Errorf("High 必须大于 Low")
		}
		turnover, err := parseFloatField(k.TurnoverRate, "TurnoverRate")
		if err != nil {
			return nil, err
		}
		if turnover < 0 {
			return nil, fmt.Errorf("TurnoverRate 必须大于等于 0")
		}
		parsed[i] = parsedKLine{
			day:      k.Day,
			open:     open,
			close:    close,
			high:     high,
			low:      low,
			turnover: turnover,
		}
		if low < minPrice {
			minPrice = low
		}
		if high > maxPrice {
			maxPrice = high
		}
	}

	if minPrice == math.MaxFloat64 {
		return nil, fmt.Errorf("无法解析最低价")
	}
	span := maxPrice - minPrice
	if span <= 0 {
		span = math.Max(maxPrice*0.01, 0.01)
		minPrice -= span / 4
		maxPrice += span / 4
		span = maxPrice - minPrice
	}

	bucketCount := accuracyFactor
	if bucketCount < 3 {
		bucketCount = 3
	}
	bucketWidth := span / float64(bucketCount-1)
	if bucketWidth <= 0 {
		bucketWidth = 0.01
	}
	centers := make([]float64, bucketCount)
	for i := 0; i < bucketCount; i++ {
		centers[i] = minPrice + bucketWidth*float64(i)
	}

	weights := make([]float64, bucketCount)
	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].day < parsed[j].day
	})
	initialized := false
	for _, day := range parsed {
		turnoverRate := math.Min(1.0, math.Max(0.0, day.turnover/100))
		if initialized {
			for i := range weights {
				weights[i] *= (1 - turnoverRate)
			}
		}
		newWeight := turnoverRate
		if !initialized {
			newWeight = 1
		}
		pdf := buildTriangularPDF(centers, day.low, day.high, averagePrice(day))
		sumPDF := sumSlice(pdf)
		if sumPDF <= 0 {
			singleIdx := bestCenterIndex(centers, averagePrice(day))
			pdf = make([]float64, bucketCount)
			pdf[singleIdx] = 1
			sumPDF = 1
		}
		for i := range weights {
			weights[i] += newWeight * pdf[i] / sumPDF
		}
		normalize(weights)
		initialized = true
	}

	if !initialized {
		return nil, fmt.Errorf("筹码分布计算失败")
	}

	distribution := make([]DistributionBucket, bucketCount)
	for i := 0; i < bucketCount; i++ {
		distribution[i] = DistributionBucket{
			Price:  centers[i],
			Weight: weights[i],
		}
	}

	latest := parsed[len(parsed)-1]
	benefit := 0.0
	avgCost := 0.0
	for i, center := range centers {
		avgCost += center * weights[i]
		if center <= latest.close {
			benefit += weights[i]
		}
	}
	benefitPart := math.Max(0, math.Min(benefit*100, 100))
	cost70Lower, cost70Upper := intervalForCoverage(centers, weights, 0.7)
	cost90Lower, cost90Upper := intervalForCoverage(centers, weights, 0.9)

	return &ChipDistributionResult{
		LatestDate:  latest.day,
		LatestClose: latest.close,
		SampleDays:  len(parsed),
		BenefitPart: benefitPart,
		AvgCost:     avgCost,
		Cost70: DistributionInterval{
			LowerPrice:    cost70Lower,
			UpperPrice:    cost70Upper,
			Concentration: calculateConcentration(cost70Lower, cost70Upper),
		},
		Cost90: DistributionInterval{
			LowerPrice:    cost90Lower,
			UpperPrice:    cost90Upper,
			Concentration: calculateConcentration(cost90Lower, cost90Upper),
		},
		Distribution: distribution,
	}, nil
}

func parseFloatField(value, name string) (float64, error) {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, "%")
	if trimmed == "" {
		return 0, fmt.Errorf("%s 为空", name)
	}
	f, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("%s 解析失败: %w", name, err)
	}
	return f, nil
}

func averagePrice(day parsedKLine) float64 {
	return (day.open + day.close + day.high + day.low) / 4
}

func buildTriangularPDF(centers []float64, low, high, peak float64) []float64 {
	pdf := make([]float64, len(centers))
	if high < low {
		return pdf
	}
	peak = math.Max(low, math.Min(high, peak))
	if high == low {
		pdf[bestCenterIndex(centers, high)] = 1
		return pdf
	}
	for i, c := range centers {
		if c < low || c > high {
			continue
		}
		if c <= peak {
			if peak == low {
				pdf[i] = 1
				continue
			}
			pdf[i] = (c - low) / (peak - low)
		} else {
			if high == peak {
				pdf[i] = 1
				continue
			}
			pdf[i] = (high - c) / (high - peak)
		}
	}
	return pdf
}

func sumSlice(nums []float64) float64 {
	sum := 0.0
	for _, v := range nums {
		sum += v
	}
	return sum
}

func normalize(nums []float64) {
	sum := sumSlice(nums)
	if sum == 0 {
		return
	}
	for i := range nums {
		nums[i] /= sum
	}
}

func bestCenterIndex(centers []float64, target float64) int {
	best := 0
	bestDiff := math.MaxFloat64
	for i, c := range centers {
		delta := math.Abs(c - target)
		if delta < bestDiff {
			bestDiff = delta
			best = i
		}
	}
	return best
}

func intervalForCoverage(centers, weights []float64, target float64) (float64, float64) {
	n := len(weights)
	bestWidth := math.MaxFloat64
	bestLow := centers[0]
	bestHigh := centers[n-1]
	for start := 0; start < n; start++ {
		sum := 0.0
		for end := start; end < n; end++ {
			sum += weights[end]
			if sum >= target {
				lower := centers[start]
				upper := centers[end]
				width := upper - lower
				if width < bestWidth {
					bestWidth = width
					bestLow = lower
					bestHigh = upper
				}
				break
			}
		}
	}
	return bestLow, bestHigh
}

func calculateConcentration(lower, upper float64) float64 {
	if upper == 0 {
		return 0
	}
	return ((upper - lower) / upper) * 100
}
