package app

import (
	"fmt"
	"time"

	"github.com/kyc001/paper-radar/internal/digest"
	"github.com/kyc001/paper-radar/internal/scoring"
	"github.com/kyc001/paper-radar/internal/state"
)

type DigestOptions struct {
	StatePath string
	OutputDir string
	Date      time.Time
	TopN      int
}

func RunDigest(opts DigestOptions) (string, int, error) {
	store := state.New(defaultStatePath(opts.StatePath))
	st, err := store.Load()
	if err != nil {
		return "", 0, fmt.Errorf("load state: %w", err)
	}

	scoring.SortByScore(st.Pending)

	target := st.Pending
	if opts.TopN > 0 && len(target) > opts.TopN {
		target = st.Pending[:opts.TopN]
	}

	outputPath, err := digest.WriteDaily(defaultOutputDir(opts.OutputDir), opts.Date, target)
	if err != nil {
		return "", 0, fmt.Errorf("write digest: %w", err)
	}

	count := len(target)
	if count >= len(st.Pending) {
		st.Pending = nil
	} else {
		st.Pending = st.Pending[count:]
	}

	if err := store.Save(st); err != nil {
		return "", 0, fmt.Errorf("save state: %w", err)
	}

	return outputPath, count, nil
}

func defaultOutputDir(path string) string {
	if path != "" {
		return path
	}
	return "outputs"
}
