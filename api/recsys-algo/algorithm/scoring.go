package algorithm

import "math"

const similarityTieEpsilon = 1e-9

// ApplyBlendedScoring computes normalized signals and the blended score.
func ApplyBlendedScoring(data *CandidateData, weights BlendWeights) {
	if data == nil {
		return
	}

	for i := range data.Candidates {
		id := data.Candidates[i].ItemID

		popRaw := data.PopScores[id]
		popNorm := normalizePositiveScore(popRaw)

		coocRaw := data.CoocScores[id]
		coocNorm := normalizePositiveScore(coocRaw)

		simRaw, simNorm, simSources := selectSimilarity(data, id)

		data.PopNorm[id] = popNorm
		data.CoocNorm[id] = coocNorm
		data.SimilarityNorm[id] = simNorm
		data.PopRaw[id] = popRaw
		data.CoocRaw[id] = coocRaw
		data.SimilarityRaw[id] = simRaw
		if len(simSources) > 0 {
			data.SimilaritySources[id] = simSources
		}

		blended := weights.Pop*popNorm + weights.Cooc*coocNorm + weights.Similarity*simNorm
		data.Candidates[i].Score = blended
	}
}

func selectSimilarity(data *CandidateData, id string) (float64, float64, []Signal) {
	embRaw := data.EmbScores[id]
	collabRaw := data.CollabScores[id]
	contentRaw := data.ContentScores[id]
	sessionRaw := data.SessionScores[id]

	embNorm := normalizeEmbeddingScore(embRaw)
	collabNorm := normalizePositiveScore(collabRaw)
	contentNorm := normalizePositiveScore(contentRaw)
	sessionNorm := normalizePositiveScore(sessionRaw)

	maxNorm := maxFloat(embNorm, collabNorm, contentNorm, sessionNorm)
	if maxNorm <= 0 {
		return 0, 0, nil
	}

	entries := []struct {
		signal Signal
		raw    float64
		norm   float64
	}{
		{SignalEmbedding, embRaw, embNorm},
		{SignalCollaborative, collabRaw, collabNorm},
		{SignalContent, contentRaw, contentNorm},
		{SignalSession, sessionRaw, sessionNorm},
	}

	sources := make([]Signal, 0, len(entries))
	raw := 0.0
	for _, entry := range entries {
		if entry.norm <= 0 {
			continue
		}
		if math.Abs(entry.norm-maxNorm) <= similarityTieEpsilon {
			sources = append(sources, entry.signal)
			if raw == 0 {
				raw = entry.raw
			}
		}
	}

	return raw, maxNorm, sources
}

func normalizeEmbeddingScore(score float64) float64 {
	if score <= 0 {
		return 0
	}
	if score >= 1 {
		return 1
	}
	return score
}

func normalizePositiveScore(score float64) float64 {
	if score <= 0 {
		return 0
	}
	return score / (score + 1)
}

func maxFloat(values ...float64) float64 {
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}
