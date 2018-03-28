package functions_test

import (
	"math"
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestPercentileOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"percentile","kind":"percentile","spec":{"percentile":0.9}}`)
	op := &query.Operation{
		ID: "percentile",
		Spec: &functions.PercentileOpSpec{
			Percentile: 0.9,
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestPercentile_Process(t *testing.T) {
	testCases := []struct {
		name       string
		data       []float64
		percentile float64
		exact      bool
		want       float64
	}{
		{
			name:       "zero",
			data:       []float64{0, 0, 0},
			percentile: 0.5,
			want:       0.0,
		},
		{
			name:       "50th",
			data:       []float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1},
			percentile: 0.5,
			want:       3,
		},
		{
			name:       "75th",
			data:       []float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1},
			percentile: 0.75,
			want:       4,
		},
		{
			name:       "90th",
			data:       []float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1},
			percentile: 0.9,
			want:       5,
		},
		{
			name:       "99th",
			data:       []float64{1, 2, 3, 4, 5, 5, 4, 3, 2, 1},
			percentile: 0.99,
			want:       5,
		},
		{
			name:       "exact 50th",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 0.5,
			exact:      true,
			want:       3,
		},
		{
			name:       "exact 75th",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 0.75,
			exact:      true,
			want:       4,
		},
		{
			name:       "exact 90th",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 0.9,
			exact:      true,
			want:       4.6,
		},
		{
			name:       "exact 99th",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 0.99,
			exact:      true,
			want:       4.96,
		},
		{
			name:       "exact 100th",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 1,
			exact:      true,
			want:       5,
		},
		{
			name:       "exact 50th normal",
			data:       NormalData,
			percentile: 0.5,
			exact:      true,
			want:       9.997645059676595,
		},
		{
			name:       "normal",
			data:       NormalData,
			percentile: 0.9,
			want:       13.843815760607427,
		},
		{
			name: "NaN",
			data: []float64{},
			want: math.NaN(),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var agg execute.Aggregate
			if tc.exact {
				agg = &functions.ExactPercentileAgg{Quantile: tc.percentile}
			} else {
				agg = &functions.PercentileAgg{
					Quantile:    tc.percentile,
					Compression: 1000,
				}
			}
			executetest.AggFuncTestHelper(
				t,
				agg,
				tc.data,
				tc.want,
			)
		})
	}
}

func BenchmarkPercentile(b *testing.B) {
	executetest.AggFuncBenchmarkHelper(
		b,
		&functions.PercentileAgg{
			Quantile:    0.9,
			Compression: 1000,
		},
		NormalData,
		13.843815760607427,
	)
}
