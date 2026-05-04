// Package diff implements a tiny line-based diff via Hunt-McIlroy LCS.
//
// Targeted at run-log comparisons, where input sizes are typically a few
// thousand lines. O(n*m) memory; fine at this scale, would need patching
// for log-volume use.
package diff

type Op int

const (
	Equal Op = iota
	Insert
	Delete
)

type Hunk struct {
	Op   Op     `json:"op"` // 0=equal, 1=insert (in B but not A), 2=delete (in A but not B)
	Text string `json:"text"`
}

// Lines computes a line-level diff between two strings. Each Hunk holds a
// single line. Trailing newlines are preserved.
func Lines(a, b string) []Hunk {
	la := splitKeepNewlines(a)
	lb := splitKeepNewlines(b)
	return diffSlices(la, lb)
}

func splitKeepNewlines(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// diffSlices returns hunks via classic LCS DP. Each hunk holds one line.
func diffSlices(a, b []string) []Hunk {
	n, m := len(a), len(b)

	// Trim common prefix / suffix to keep the DP small.
	prefix := 0
	for prefix < n && prefix < m && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < n-prefix && suffix < m-prefix && a[n-1-suffix] == b[m-1-suffix] {
		suffix++
	}

	out := []Hunk{}
	for i := 0; i < prefix; i++ {
		out = append(out, Hunk{Op: Equal, Text: a[i]})
	}

	midA := a[prefix : n-suffix]
	midB := b[prefix : m-suffix]
	out = append(out, lcsDiff(midA, midB)...)

	for i := n - suffix; i < n; i++ {
		out = append(out, Hunk{Op: Equal, Text: a[i]})
	}
	return out
}

func lcsDiff(a, b []string) []Hunk {
	n, m := len(a), len(b)
	if n == 0 {
		out := make([]Hunk, m)
		for i := range b {
			out[i] = Hunk{Op: Insert, Text: b[i]}
		}
		return out
	}
	if m == 0 {
		out := make([]Hunk, n)
		for i := range a {
			out[i] = Hunk{Op: Delete, Text: a[i]}
		}
		return out
	}

	// LCS table
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Walk back, building output in reverse.
	out := []Hunk{}
	i, j := n, m
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1]:
			out = append(out, Hunk{Op: Equal, Text: a[i-1]})
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			out = append(out, Hunk{Op: Insert, Text: b[j-1]})
			j--
		default:
			out = append(out, Hunk{Op: Delete, Text: a[i-1]})
			i--
		}
	}
	// reverse
	for l, r := 0, len(out)-1; l < r; l, r = l+1, r-1 {
		out[l], out[r] = out[r], out[l]
	}
	return out
}
