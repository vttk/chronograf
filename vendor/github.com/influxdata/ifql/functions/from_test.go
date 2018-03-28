package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/querytest"
)

func TestFrom_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from",
			Raw:     `from()`,
			WantErr: true,
		},
		{
			Name:    "from",
			Raw:     `from(db:"telegraf", db:"oops")`,
			WantErr: true,
		},
		{
			Name:    "from",
			Raw:     `from(db:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "from with database",
			Raw:  `from(db:"mydb") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
						},
					},
					{
						ID:   "sum2",
						Spec: &functions.SumOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestFromOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"from","kind":"from","spec":{"database":"mydb"}}`)
	op := &query.Operation{
		ID: "from",
		Spec: &functions.FromOpSpec{
			Database: "mydb",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}
