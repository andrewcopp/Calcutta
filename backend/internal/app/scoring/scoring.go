package scoring

type Rule struct {
	WinIndex      int
	PointsAwarded int
}

func PointsForProgress(rules []Rule, wins int, byes int) int {
	p := wins + byes
	if p <= 0 {
		return 0
	}
	pts := 0
	for _, r := range rules {
		if r.WinIndex <= p {
			pts += r.PointsAwarded
		}
	}
	return pts
}
