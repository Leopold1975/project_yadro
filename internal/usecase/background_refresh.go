package usecase

import (
	"context"
	"time"

	"github.com/Leopold1975/yadro_app/pkg/logger"
)

type BackgroundRefreshUsecase struct {
	fetch       FetchComicsUsecase
	refreshTime time.Time
}

func NewBackgroundRefresh(fetch FetchComicsUsecase, refreshTime time.Time) BackgroundRefreshUsecase {
	return BackgroundRefreshUsecase{
		fetch:       fetch,
		refreshTime: refreshTime,
	}
}

// Refresh не запускает обновление базы комиксов при запуске.
func (b BackgroundRefreshUsecase) Refresh(ctx context.Context, l logger.Logger) {
	now := time.Now()
	refreshTime := time.Date(now.Year(), now.Month(), now.Day(),
		b.refreshTime.Hour(), b.refreshTime.Minute(), b.refreshTime.Second(), 0, b.refreshTime.Location())

	if refreshTime.Before(now) {
		refreshTime = refreshTime.AddDate(0, 0, 1)
	}

	timer := time.NewTimer(refreshTime.Sub(now))
	defer timer.Stop()

	l.Info("Start background refresh")

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			resp, err := b.fetch.FetchComics(ctx)
			if err != nil {
				l.Error("background refresh error", "error", err)

				continue
			}

			l.Info("refreshed", "new comics", resp.New, "total comics", resp.Total)

			refreshTime = b.calculateNextRefreshTime(refreshTime)
			timer.Reset(time.Until(refreshTime))
		}
	}
}

func (b BackgroundRefreshUsecase) calculateNextRefreshTime(refreshTime time.Time) time.Time {
	return refreshTime.Add(24 * time.Hour) //nolint:gomnd
}
