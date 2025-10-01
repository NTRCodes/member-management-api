package app

import (
	"NTRCodes/member-api/internal/database"
	"NTRCodes/member-api/internal/database/members"
	"NTRCodes/member-api/internal/database/metrics"
	"context"
	"os"
)

type App struct {
	RequiredEnv  []string
	DB           *database.DB
	MemberRepo   members.MemberRepository
	MetricsRepo  metrics.MetricsRepository
}

func New(memberRepo members.MemberRepository, metricsRepo metrics.MetricsRepository) *App {
	return &App{
		RequiredEnv: []string{"PORT", "DB_DSN"}, // add more later: MEMBER_SSN_SALT, DB_DSN
		MemberRepo:  memberRepo,
		MetricsRepo: metricsRepo,
	}
}

func (a *App) Ready(ctx context.Context) (bool, string) {
	for _, name := range a.RequiredEnv {
		if os.Getenv(name) == "" {
			return false, "missing env: " + name
		}
	}
	if a.DB == nil {
		return false, "db not initialized"
	}
	if err := a.DB.HealthCheck(ctx); err != nil {
		return false, "db not ready: " + err.Error()
	}
	return true, "ready"
}
