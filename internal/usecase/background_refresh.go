package usecase

import (
	"context"
	"time"

	"github.com/Leopold1975/yadro_app/pkg/logger"
)

type BackgroundRefreshUsecase struct {
	fetch    FetchComicsUsecase
	interval time.Duration
}

func NewBackgroundRefresh(fetch FetchComicsUsecase, interval time.Duration) BackgroundRefreshUsecase {
	return BackgroundRefreshUsecase{
		fetch:    fetch,
		interval: interval,
	}
}

// Refresh не запускает обновление базы комиксов при запуске.
func (b BackgroundRefreshUsecase) Refresh(ctx context.Context, l logger.Logger) {
	ticker := time.NewTicker(b.interval)

	l.Info("Start background refresh")
	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		_, err := b.fetch.FetchComics(ctx)
		l.Error("background refrsh error", err)
	}
}
