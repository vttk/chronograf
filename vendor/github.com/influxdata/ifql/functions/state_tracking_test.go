package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
	"github.com/influxdata/ifql/semantic"
)

func TestStateTrackingOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"id","kind":"stateTracking","spec":{"count_label":"c","duration_label":"d","duration_unit":"1m"}}`)
	op := &query.Operation{
		ID: "id",
		Spec: &functions.StateTrackingOpSpec{
			CountLabel:    "c",
			DurationLabel: "d",
			DurationUnit:  query.Duration(time.Minute),
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestStateTracking_Process(t *testing.T) {
	gt5 := &semantic.FunctionExpression{
		Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "r"}}},
		Body: &semantic.BinaryExpression{
			Operator: ast.GreaterThanOperator,
			Left: &semantic.MemberExpression{
				Object:   &semantic.IdentifierExpression{Name: "r"},
				Property: "_value",
			},
			Right: &semantic.FloatLiteral{Value: 5.0},
		},
	}
	testCases := []struct {
		name string
		spec *functions.StateTrackingProcedureSpec
		data []execute.Block
		want []*executetest.Block
	}{
		{
			name: "one block",
			spec: &functions.StateTrackingProcedureSpec{
				CountLabel:    "count",
				DurationLabel: "duration",
				DurationUnit:  1,
				Fn:            gt5,
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{execute.Time(4), 7.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "count", Type: execute.TInt, Kind: execute.ValueColKind},
					{Label: "duration", Type: execute.TInt, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(-1), int64(-1)},
					{execute.Time(2), 1.0, int64(-1), int64(-1)},
					{execute.Time(3), 6.0, int64(1), int64(0)},
					{execute.Time(4), 7.0, int64(2), int64(1)},
					{execute.Time(5), 8.0, int64(3), int64(2)},
					{execute.Time(6), 1.0, int64(-1), int64(-1)},
				},
			}},
		},
		{
			name: "only duration",
			spec: &functions.StateTrackingProcedureSpec{
				DurationLabel: "duration",
				DurationUnit:  1,
				Fn:            gt5,
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{execute.Time(4), 7.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "duration", Type: execute.TInt, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(-1)},
					{execute.Time(2), 1.0, int64(-1)},
					{execute.Time(3), 6.0, int64(0)},
					{execute.Time(4), 7.0, int64(1)},
					{execute.Time(5), 8.0, int64(2)},
					{execute.Time(6), 1.0, int64(-1)},
				},
			}},
		},
		{
			name: "only count",
			spec: &functions.StateTrackingProcedureSpec{
				CountLabel: "count",
				Fn:         gt5,
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0},
					{execute.Time(2), 1.0},
					{execute.Time(3), 6.0},
					{execute.Time(4), 7.0},
					{execute.Time(5), 8.0},
					{execute.Time(6), 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  3,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "count", Type: execute.TInt, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(1), 2.0, int64(-1)},
					{execute.Time(2), 1.0, int64(-1)},
					{execute.Time(3), 6.0, int64(1)},
					{execute.Time(4), 7.0, int64(2)},
					{execute.Time(5), 8.0, int64(3)},
					{execute.Time(6), 1.0, int64(-1)},
				},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
					tx, err := functions.NewStateTrackingTransformation(d, c, tc.spec)
					if err != nil {
						t.Fatal(err)
					}
					return tx
				},
			)
		})
	}
}
