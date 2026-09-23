package services

import (
	"context"
	"log"
	"time"

	"kalasetu/repos"
)

// CheckoutSessionSweeper periodically releases every expired Checkout
// Session, so an abandoned Buyer's Stock becomes buyable again on browse
// pages without anyone attempting a purchase or confirmation.
type CheckoutSessionSweeper struct {
	repo     repos.CheckoutSessionRepository
	interval time.Duration
}

func NewCheckoutSessionSweeper(repo repos.CheckoutSessionRepository, interval time.Duration) *CheckoutSessionSweeper {
	return &CheckoutSessionSweeper{repo: repo, interval: interval}
}

// Start runs the sweep on a ticker until ctx is cancelled. It blocks, so
// callers run it in its own goroutine.
func (s *CheckoutSessionSweeper) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.repo.ReleaseExpired(ctx); err != nil {
				log.Printf("checkout session sweeper: %v", err)
			}
		}
	}
}
