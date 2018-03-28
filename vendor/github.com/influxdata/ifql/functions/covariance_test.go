package functions_test

import (
	"testing"

	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
	"github.com/influxdata/ifql/semantic"
)

func TestCovariance_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "simple covariance",
			Raw:  `from(db:"mydb") |> covariance()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID:   "covariance1",
						Spec: &functions.CovarianceOpSpec{},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "covariance1"},
				},
			},
		},
		{
			Name: "pearsonr",
			Raw:  `from(db:"mydb")|>covariance(pearsonr:true)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "covariance1",
						Spec: &functions.CovarianceOpSpec{
							PearsonCorrelation: true,
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "covariance1"},
				},
			},
		},
		{
			Name: "global covariance",
			Raw:  `cov(x: from(db:"mydb"), y:from(db:"mydb"), on:["host"], pearsonr:true)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "from1",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "join2",
						Spec: &functions.JoinOpSpec{
							On: []string{"host"},
							TableNames: map[query.OperationID]string{
								"from0": "x",
								"from1": "y",
							},
							Fn: &semantic.FunctionExpression{
								Params: []*semantic.FunctionParam{
									{Key: &semantic.Identifier{Name: "t"}},
								},
								Body: &semantic.ObjectExpression{
									Properties: []*semantic.Property{
										{
											Key: &semantic.Identifier{Name: "x"},
											Value: &semantic.MemberExpression{
												Object: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "t"},
													Property: "x",
												},
												Property: "_value",
											},
										},
										{
											Key: &semantic.Identifier{Name: "y"},
											Value: &semantic.MemberExpression{
												Object: &semantic.MemberExpression{
													Object:   &semantic.IdentifierExpression{Name: "t"},
													Property: "y",
												},
												Property: "_value",
											},
										},
									},
								},
							},
						},
					},
					{
						ID: "covariance3",
						Spec: &functions.CovarianceOpSpec{
							PearsonCorrelation: true,
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "join2"},
					{Parent: "from1", Child: "join2"},
					{Parent: "join2", Child: "covariance3"},
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

func TestCovarianceOperation_Marshaling(t *testing.T) {
	data := []byte(`{
		"id":"covariance",
		"kind":"covariance",
		"spec":{
			"pearsonr":true
		}
	}`)
	op := &query.Operation{
		ID: "covariance",
		Spec: &functions.CovarianceOpSpec{
			PearsonCorrelation: true,
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestCovariance_Process(t *testing.T) {
	testCases := []struct {
		name   string
		bounds execute.Bounds
		spec   *functions.CovarianceProcedureSpec
		data   []execute.Block
		want   []*executetest.Block
	}{
		{
			name: "variance",
			bounds: execute.Bounds{
				Start: 1,
				Stop:  5,
			},
			spec: &functions.CovarianceProcedureSpec{},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "x", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "y", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, 1.0},
					{execute.Time(1), 2.0, 2.0},
					{execute.Time(2), 3.0, 3.0},
					{execute.Time(3), 4.0, 4.0},
					{execute.Time(4), 5.0, 5.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(5), 2.5},
				},
			}},
		},
		{
			name: "negative covariance",
			bounds: execute.Bounds{
				Start: 1,
				Stop:  5,
			},
			spec: &functions.CovarianceProcedureSpec{},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "x", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "y", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, 5.0},
					{execute.Time(1), 2.0, 4.0},
					{execute.Time(2), 3.0, 3.0},
					{execute.Time(3), 4.0, 2.0},
					{execute.Time(4), 5.0, 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(5), -2.5},
				},
			}},
		},
		{
			name: "small covariance",
			bounds: execute.Bounds{
				Start: 1,
				Stop:  5,
			},
			spec: &functions.CovarianceProcedureSpec{},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "x", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "y", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, 1.0},
					{execute.Time(1), 2.0, 1.0},
					{execute.Time(2), 3.0, 1.0},
					{execute.Time(3), 4.0, 1.0},
					{execute.Time(4), 5.0, 2.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(5), 0.5},
				},
			}},
		},
		{
			name: "pearson correlation",
			bounds: execute.Bounds{
				Start: 1,
				Stop:  5,
			},
			spec: &functions.CovarianceProcedureSpec{
				PearsonCorrelation: true,
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "x", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "y", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, 1.0},
					{execute.Time(1), 2.0, 2.0},
					{execute.Time(2), 3.0, 3.0},
					{execute.Time(3), 4.0, 4.0},
					{execute.Time(4), 5.0, 5.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(5), 1.0},
				},
			}},
		},
		{
			name: "pearson correlation opposite",
			bounds: execute.Bounds{
				Start: 1,
				Stop:  5,
			},
			spec: &functions.CovarianceProcedureSpec{
				PearsonCorrelation: true,
			},
			data: []execute.Block{&executetest.Block{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "x", Type: execute.TFloat, Kind: execute.ValueColKind},
					{Label: "y", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(0), 1.0, 5.0},
					{execute.Time(1), 2.0, 4.0},
					{execute.Time(2), 3.0, 3.0},
					{execute.Time(3), 4.0, 2.0},
					{execute.Time(4), 5.0, 1.0},
				},
			}},
			want: []*executetest.Block{{
				Bnds: execute.Bounds{
					Start: 1,
					Stop:  5,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
				},
				Data: [][]interface{}{
					{execute.Time(5), -1.0},
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
					return functions.NewCovarianceTransformation(d, c, tc.spec, tc.bounds)
				},
			)
		})
	}
}
