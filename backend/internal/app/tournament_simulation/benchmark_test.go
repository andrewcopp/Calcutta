package tournament_simulation

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func BenchmarkSimulateToyBracket(b *testing.B) {
	br := toyBracket()
	probs := map[MatchupKey]float64{
		{GameID: "g1", Team1ID: "t1", Team2ID: "t3"}: 0.6,
		{GameID: "g2", Team1ID: "t2", Team2ID: "t4"}: 0.7,

		{GameID: "g3", Team1ID: "t1", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t1", Team2ID: "t4"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t4"}: 0.5,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Simulate(br, probs, 5000, 42, Options{Workers: 1})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSimulateToyBracketParallel(b *testing.B) {
	br := toyBracket()
	probs := map[MatchupKey]float64{
		{GameID: "g1", Team1ID: "t1", Team2ID: "t3"}: 0.6,
		{GameID: "g2", Team1ID: "t2", Team2ID: "t4"}: 0.7,

		{GameID: "g3", Team1ID: "t1", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t1", Team2ID: "t4"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t2"}: 0.5,
		{GameID: "g3", Team1ID: "t3", Team2ID: "t4"}: 0.5,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Simulate(br, probs, 5000, 42, Options{Workers: 8})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundOrder(b *testing.B) {
	rounds := []models.BracketRound{
		models.RoundFirstFour,
		models.RoundOf64,
		models.RoundOf32,
		models.RoundSweet16,
		models.RoundElite8,
		models.RoundFinalFour,
		models.RoundChampionship,
		"",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = rounds[i%len(rounds)].Order()
	}
}
