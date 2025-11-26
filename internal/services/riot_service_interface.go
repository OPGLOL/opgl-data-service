package services

import "github.com/OPGLOL/opgl-data-service/internal/models"

// RiotServiceInterface defines the interface for Riot API operations
// This allows for easy mocking in tests
type RiotServiceInterface interface {
	GetSummonerByRiotID(region string, gameName string, tagLine string) (*models.Summoner, error)
	GetSummonerByPUUID(region string, puuid string) (*models.Summoner, error)
	GetMatchHistory(region string, puuid string, count int) ([]models.Match, error)
	GetMatchDetails(region string, matchID string) (*models.Match, error)
}

// Verify RiotService implements RiotServiceInterface
var _ RiotServiceInterface = (*RiotService)(nil)
