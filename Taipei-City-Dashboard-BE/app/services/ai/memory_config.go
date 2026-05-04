package ai

import "TaipeiCityDashboardBE/global"

func configuredMemory() memoryConfig {
	turns := clampInt(global.TWCC.MemoryTurns, minMemoryTurns, maxMemoryTurns)
	blockRunes := clampInt(global.TWCC.MemoryBlockRunes, minMemoryBlockRunes, maxMemoryBlockRunes)
	return memoryConfig{
		maxTurns:       turns,
		blockRunes:     blockRunes,
		perTurnRunes:   clampInt(blockRunes/turns, minMemoryTurnRunes, maxMemoryTurnRunes),
		candidateLimit: turns * 2,
	}
}

func newMemoryStats(cfg memoryConfig) memoryStats {
	return memoryStats{Source: memorySourceAIChatLog, MaxTurns: cfg.maxTurns}
}

func (s *aiSession) metadataParams() map[string]interface{} {
	params := make(map[string]interface{}, len(s.req.Params)+6)
	for key, value := range s.req.Params {
		params[key] = value
	}
	stats := s.memoryStats
	if stats.Source == "" {
		stats = newMemoryStats(configuredMemory())
	}
	params["memory_loaded"] = stats.Loaded
	params["memory_source"] = stats.Source
	params["memory_turns_loaded"] = stats.TurnsLoaded
	params["memory_max_turns"] = stats.MaxTurns
	params["memory_truncated"] = stats.Truncated
	params["memory_block_runes"] = stats.BlockRunes
	return params
}

func currentRequestRunes(req AIChatRequest) int {
	total := 0
	for _, msg := range req.Messages {
		total += len([]rune(extractText(msg)))
	}
	return total
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
