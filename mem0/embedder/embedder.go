package embedder

import (
	"context"
	"crypto/sha256"
	"errors"
	"math"
	"strings"
)

type Embedder interface {
	Embed(ctx context.Context, input []string) ([][]float64, error)
	Dim() int
}

type HashEmbedder struct {
	dim int
}

func NewHashEmbedder(dim int) (*HashEmbedder, error) {
	if dim <= 0 {
		return nil, errors.New("dim must be > 0")
	}
	return &HashEmbedder{dim: dim}, nil
}

func (h *HashEmbedder) Dim() int { return h.dim }

func (h *HashEmbedder) Embed(ctx context.Context, input []string) ([][]float64, error) {
	_ = ctx
	out := make([][]float64, 0, len(input))
	for _, s := range input {
		vec := make([]float64, h.dim)
		tokens := tokenize(s)
		if len(tokens) == 0 {
			out = append(out, vec)
			continue
		}
		for _, tok := range tokens {
			sum := sha256.Sum256([]byte(tok))
			idx := int(sum[0])<<8 | int(sum[1])
			sign := 1.0
			if sum[2]&1 == 1 {
				sign = -1.0
			}
			vec[idx%h.dim] += sign
		}
		n := l2norm(vec)
		if n > 0 {
			for i := range vec {
				vec[i] /= n
			}
		}
		out = append(out, vec)
	}
	return out, nil
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer(",", " ", "，", " ", ":", " ", "：", " ", "(", " ", ")", " ", "[", " ", "]", " ", "{", " ", "}", " ", "\"", " ", "'", " ", "\t", " ")
	s = repl.Replace(s)
	fields := strings.Fields(s)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if len([]rune(f)) < 2 {
			continue
		}
		out = append(out, f)
	}
	return out
}

func l2norm(v []float64) float64 {
	var s float64
	for _, x := range v {
		s += x * x
	}
	return math.Sqrt(s)
}

func Cosine(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dot, na, nb float64
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
