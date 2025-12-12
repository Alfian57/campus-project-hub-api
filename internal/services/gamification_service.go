package services

// Gamification constants - EXP points for actions
const (
	ExpCreateProject  = 100
	ExpSellProject    = 150
	ExpBuyProject     = 50
	ExpReceiveLike    = 10
	ExpReceiveComment = 5
	ExpProjectViewed  = 1
	ExpCreateArticle  = 75
	ExpArticleViewed  = 1
)

// Level configuration
const (
	BaseExp  = 100
	MaxLevel = 100
)

// GamificationConfig for API response
type GamificationConfig struct {
	ActionPoints map[string]int `json:"actionPoints"`
	LevelConfig  LevelConfig    `json:"levelConfig"`
	LevelTitles  map[int]string `json:"levelTitles"`
}

type LevelConfig struct {
	BaseExp    int `json:"baseExp"`
	Multiplier int `json:"multiplier"`
	MaxLevel   int `json:"maxLevel"`
}

// GetGamificationConfig returns the gamification configuration
func GetGamificationConfig() GamificationConfig {
	return GamificationConfig{
		ActionPoints: map[string]int{
			"CREATE_PROJECT":  ExpCreateProject,
			"SELL_PROJECT":    ExpSellProject,
			"BUY_PROJECT":     ExpBuyProject,
			"RECEIVE_LIKE":    ExpReceiveLike,
			"RECEIVE_COMMENT": ExpReceiveComment,
			"PROJECT_VIEWED":  ExpProjectViewed,
			"CREATE_ARTICLE":  ExpCreateArticle,
			"ARTICLE_VIEWED":  ExpArticleViewed,
		},
		LevelConfig: LevelConfig{
			BaseExp:    BaseExp,
			Multiplier: 2,
			MaxLevel:   MaxLevel,
		},
		LevelTitles: map[int]string{
			1:   "Pemula",
			5:   "Aktif",
			10:  "Contributor",
			20:  "Expert",
			50:  "Master",
			100: "Legend",
		},
	}
}

// GetLevelFromExp calculates level from total EXP
func GetLevelFromExp(totalExp int) int {
	level := 1
	for level < MaxLevel {
		required := BaseExp * level * level
		if totalExp < required {
			break
		}
		level++
	}
	return level
}

// GetRequiredExpForLevel calculates required EXP for a specific level
func GetRequiredExpForLevel(level int) int {
	return BaseExp * level * level
}

// GetLevelProgress calculates progress to next level (0-100)
func GetLevelProgress(totalExp int) int {
	currentLevel := GetLevelFromExp(totalExp)
	if currentLevel >= MaxLevel {
		return 100
	}

	currentLevelExp := GetRequiredExpForLevel(currentLevel)
	nextLevelExp := GetRequiredExpForLevel(currentLevel + 1)
	expInCurrentLevel := totalExp - currentLevelExp
	expNeeded := nextLevelExp - currentLevelExp

	if expNeeded <= 0 {
		return 100
	}

	return (expInCurrentLevel * 100) / expNeeded
}

// GetExpToNextLevel calculates EXP needed for next level
func GetExpToNextLevel(totalExp int) int {
	currentLevel := GetLevelFromExp(totalExp)
	if currentLevel >= MaxLevel {
		return 0
	}

	nextLevelExp := GetRequiredExpForLevel(currentLevel + 1)
	return nextLevelExp - totalExp
}

// GetLevelTitle returns title for a given level
func GetLevelTitle(level int) string {
	titles := map[int]string{
		1:   "Pemula",
		5:   "Aktif",
		10:  "Contributor",
		20:  "Expert",
		50:  "Master",
		100: "Legend",
	}

	thresholds := []int{100, 50, 20, 10, 5, 1}
	for _, threshold := range thresholds {
		if level >= threshold {
			return titles[threshold]
		}
	}
	return "Pemula"
}

// GamificationStats for user profile
type GamificationStats struct {
	TotalExp       int    `json:"totalExp"`
	Level          int    `json:"level"`
	LevelTitle     string `json:"levelTitle"`
	LevelProgress  int    `json:"levelProgress"`
	ExpToNextLevel int    `json:"expToNextLevel"`
}

// GetUserGamificationStats returns gamification stats for a user
func GetUserGamificationStats(totalExp int) GamificationStats {
	level := GetLevelFromExp(totalExp)
	return GamificationStats{
		TotalExp:       totalExp,
		Level:          level,
		LevelTitle:     GetLevelTitle(level),
		LevelProgress:  GetLevelProgress(totalExp),
		ExpToNextLevel: GetExpToNextLevel(totalExp),
	}
}
