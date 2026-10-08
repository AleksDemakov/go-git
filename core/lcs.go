package core

import (
	"hash/fnv"
	"strings"
)

// DiffRegionType represents the type of a diff region
type DiffRegionType int

const (
	RegionCommon DiffRegionType = iota
	RegionOnlyFirst
	RegionOnlySecond
)

// DiffRegion represents a contiguous block of lines with the same diff type
type DiffRegion struct {
	Type  DiffRegionType
	Lines []string
}

// DiffFormat determines the output format for diff content
type DiffFormat int

const (
	SimpleDiffFormat DiffFormat = iota
	GitConflictFormat
)

// ComputeDiffRegions finds all diff regions between two line sequences using LCS
func ComputeDiffRegions(lines1 []string, lines2 []string) []DiffRegion {
	m := len(lines1)
	n := len(lines2)

	hashes1 := precomputeHashes(lines1)
	hashes2 := precomputeHashes(lines2)

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if hashes1[i-1] == hashes2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	var regions []DiffRegion
	i, j := m, n

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && hashes1[i-1] == hashes2[j-1] {
			if len(regions) > 0 && regions[len(regions)-1].Type == RegionCommon {
				regions[len(regions)-1].Lines = append(regions[len(regions)-1].Lines, lines1[i-1])
			} else {
				regions = append(regions, DiffRegion{Type: RegionCommon, Lines: []string{lines1[i-1]}})
			}
			i--
			j--
		} else if i > 0 && (j == 0 || dp[i-1][j] >= dp[i][j-1]) {
			if len(regions) > 0 && regions[len(regions)-1].Type == RegionOnlyFirst {
				regions[len(regions)-1].Lines = append(regions[len(regions)-1].Lines, lines1[i-1])
			} else {
				regions = append(regions, DiffRegion{Type: RegionOnlyFirst, Lines: []string{lines1[i-1]}})
			}
			i--
		} else if j > 0 {
			if len(regions) > 0 && regions[len(regions)-1].Type == RegionOnlySecond {
				regions[len(regions)-1].Lines = append(regions[len(regions)-1].Lines, lines2[j-1])
			} else {
				regions = append(regions, DiffRegion{Type: RegionOnlySecond, Lines: []string{lines2[j-1]}})
			}
			j--
		}
	}

	for left, right := 0, len(regions)-1; left < right; left, right = left+1, right-1 {
		regions[left], regions[right] = regions[right], regions[left]
	}

	return regions
}

func FormatAsSimpleDiff(regions []DiffRegion) string {
	result := ""
	for _, r := range regions {
		for _, line := range r.Lines {
			switch r.Type {
			case RegionOnlyFirst:
				result += "+" + line + "\n"
			case RegionOnlySecond:
				result += "-" + line + "\n"
			case RegionCommon:
				result += line + "\n"
			}
		}
	}
	return result
}

func FormatAsGitConflict(regions []DiffRegion, label1 string, label2 string) string {
	result := ""
	inConflict := false
	var oursLines []string
	var theirsLines []string

	for _, r := range regions {
		switch r.Type {
		case RegionCommon:
			if inConflict {
				result += "<<<<<<< " + label1 + "\n"
				for _, line := range oursLines {
					result += line + "\n"
				}
				result += "=======\n"
				for _, line := range theirsLines {
					result += line + "\n"
				}
				result += ">>>>>>> " + label2 + "\n"
				oursLines = nil
				theirsLines = nil
				inConflict = false
			}
			result += strings.Join(r.Lines, "\n") + "\n"
		case RegionOnlyFirst:
			inConflict = true
			oursLines = append(oursLines, r.Lines...)
		case RegionOnlySecond:
			inConflict = true
			theirsLines = append(theirsLines, r.Lines...)
		}
	}

	if inConflict {
		result += "<<<<<<< " + label1 + "\n"
		for _, line := range oursLines {
			result += line + "\n"
		}
		result += "=======\n"
		for _, line := range theirsLines {
			result += line + "\n"
		}
		result += ">>>>>>> " + label2 + "\n"
	}

	return result
}

func GenerateDiff(content1 string, content2 string, format DiffFormat, label1 string, label2 string) string {
	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	if content1 == content2 {
		return content1
	}

	regions := ComputeDiffRegions(lines1, lines2)

	switch format {
	case SimpleDiffFormat:
		return FormatAsSimpleDiff(regions)
	case GitConflictFormat:
		return FormatAsGitConflict(regions, label1, label2)
	default:
		return ""
	}
}

func precomputeHashes(lines []string) []uint64 {
	hashes := make([]uint64, len(lines))
	for i, line := range lines {
		hashes[i] = hashString(line)
	}
	return hashes
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
