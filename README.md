# OPGL Data Service

Microservice for retrieving League of Legends data from the Riot Games API.

## Features

- Summoner lookup by name and region
- Match history retrieval
- Health check endpoint

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Service health check |
| `/api/v1/summoner/{region}/{summonerName}` | GET | Get summoner information |
| `/api/v1/matches/{region}/{puuid}` | GET | Get match history |

## Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Add your Riot API key to .env
   ```

3. **Run the service**:
   ```bash
   go run main.go
   ```

Service runs on port **8081** by default.

## Environment Variables

- `RIOT_API_KEY` - Your Riot Games API key
- `PORT` - Service port (default: 8081)

## Testing

Use Bruno collection at `bruno-collections/opgl/opgl-data/` to test endpoints.
