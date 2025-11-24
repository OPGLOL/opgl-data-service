package models

import "time"

// Summoner represents a League of Legends player account
type Summoner struct {
	// Encrypted summoner ID returned by Riot API
	ID string `json:"id"`
	// Encrypted account ID
	AccountID string `json:"accountId"`
	// Encrypted PUUID (Player Universally Unique IDentifier)
	PUUID string `json:"puuid"`
	// Summoner name visible in game
	Name string `json:"name"`
	// Profile icon ID number
	ProfileIconID int `json:"profileIconId"`
	// Summoner level (non-ranked progression)
	SummonerLevel int64 `json:"summonerLevel"`
}

// Match represents a single League of Legends match
type Match struct {
	// Unique match identifier
	MatchID string `json:"matchId"`
	// Timestamp when the match started
	GameCreation time.Time `json:"gameCreation"`
	// Total duration of the match in seconds
	GameDuration int `json:"gameDuration"`
	// Game mode (e.g., CLASSIC, ARAM)
	GameMode string `json:"gameMode"`
	// Game type (e.g., MATCHED_GAME)
	GameType string `json:"gameType"`
	// List of all participants in the match
	Participants []Participant `json:"participants"`
}

// Participant represents a player's performance in a specific match
type Participant struct {
	// Player's PUUID
	PUUID string `json:"puuid"`
	// Summoner name at the time of the match
	SummonerName string `json:"summonerName"`
	// Champion ID played in this match
	ChampionID int `json:"championId"`
	// Champion name for easier reference
	ChampionName string `json:"championName"`
	// Number of enemy champions killed
	Kills int `json:"kills"`
	// Number of times the player died
	Deaths int `json:"deaths"`
	// Number of assists in killing enemy champions
	Assists int `json:"assists"`
	// Total gold earned during the match
	GoldEarned int `json:"goldEarned"`
	// Total damage dealt to champions
	TotalDamageDealtToChampions int `json:"totalDamageDealtToChampions"`
	// Total damage taken from all sources
	TotalDamageTaken int `json:"totalDamageTaken"`
	// Vision score (wards placed, destroyed, etc.)
	VisionScore int `json:"visionScore"`
	// Creep score (minions and monsters killed)
	TotalMinionsKilled int `json:"totalMinionsKilled"`
	// Whether the player's team won the match
	Win bool `json:"win"`
	// Player's role in the match (TOP, JUNGLE, MID, BOT, SUPPORT)
	TeamPosition string `json:"teamPosition"`
}
