package store

import (
	"context"
	"strconv"
	"testing"

	"recsys/internal/store"
	"recsys/test/shared"

	"github.com/stretchr/testify/require"
)

func TestUpsertItemAndUserFactors(t *testing.T) {
	pool := shared.MustPool(t)
	shared.CleanTables(t, pool)

	s := store.New(pool)
	orgID := shared.MustOrgID(t)
	ns := "default"

	itemVec := makeTestVector(1.0)
	require.NoError(t, s.UpsertItemFactors(context.Background(), orgID, ns, []store.ItemFactorUpsert{
		{ItemID: "item-1", Factors: itemVec},
	}))

	var storedItem string
	require.NoError(t, pool.QueryRow(context.Background(), `
        SELECT factors::text
        FROM recsys_item_factors
        WHERE org_id = $1 AND namespace = $2 AND item_id = $3
    `, orgID, ns, "item-1").Scan(&storedItem))
	require.Equal(t, literal(itemVec), storedItem)

	userVec := makeTestVector(0.5)
	require.NoError(t, s.UpsertUserFactors(context.Background(), orgID, ns, []store.UserFactorUpsert{
		{UserID: "user-1", Factors: userVec},
	}))

	var storedUser string
	require.NoError(t, pool.QueryRow(context.Background(), `
        SELECT factors::text
        FROM recsys_user_factors
        WHERE org_id = $1 AND namespace = $2 AND user_id = $3
    `, orgID, ns, "user-1").Scan(&storedUser))
	require.Equal(t, literal(userVec), storedUser)

	err := s.UpsertItemFactors(context.Background(), orgID, ns, []store.ItemFactorUpsert{
		{ItemID: "bad", Factors: []float64{1, 2, 3}},
	})
	require.Error(t, err, "expected vector length validation to trigger")
}

func makeTestVector(seed float64) []float64 {
	vec := make([]float64, store.EmbeddingDims)
	for i := range vec {
		vec[i] = seed + float64(i%7)
	}
	return vec
}

func literal(vec []float64) string {
	out := make([]byte, 0, len(vec)*8)
	out = append(out, '[')
	for i, v := range vec {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, []byte(strconv.FormatFloat(v, 'g', -1, 64))...)
	}
	out = append(out, ']')
	return string(out)
}
