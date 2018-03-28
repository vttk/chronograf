package functions_test

import (
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/functions"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
	"github.com/influxdata/ifql/query/querytest"
)

func TestWindow_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with window",
			Raw:  `from(db:"mydb") |> window(start:-4h, every:1h)`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "window1",
						Spec: &functions.WindowOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Every:  query.Duration(time.Hour),
							Period: query.Duration(time.Hour),
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "from0", Child: "window1"},
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

func TestWindowOperation_Marshaling(t *testing.T) {
	//TODO: Test marshalling of triggerspec
	data := []byte(`{"id":"window","kind":"window","spec":{"every":"1m","period":"1h","start":"-4h","round":"1s"}}`)
	op := &query.Operation{
		ID: "window",
		Spec: &functions.WindowOpSpec{
			Every:  query.Duration(time.Minute),
			Period: query.Duration(time.Hour),
			Start: query.Time{
				Relative:   -4 * time.Hour,
				IsRelative: true,
			},
			Round: query.Duration(time.Second),
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFixedWindow_PassThrough(t *testing.T) {
	executetest.TransformationPassThroughTestHelper(t, func(d execute.Dataset, c execute.BlockBuilderCache) execute.Transformation {
		fw := functions.NewFixedWindowTransformation(
			d,
			c,
			execute.Bounds{},
			execute.Window{
				Every:  execute.Duration(time.Minute),
				Period: execute.Duration(time.Minute),
			},
		)
		return fw
	})
}

func TestFixedWindow_Process(t *testing.T) {
	testCases := []struct {
		name          string
		valueCol      execute.ColMeta
		start         execute.Time
		every, period execute.Duration
		num           int
		want          func(start execute.Time) []*executetest.Block
	}{
		{
			name:     "nonoverlapping_nonaligned",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
			// Use a time that is *not* aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, 0.0},
							{start + execute.Time(10*time.Second), 1.0},
							{start + execute.Time(20*time.Second), 2.0},
							{start + execute.Time(30*time.Second), 3.0},
							{start + execute.Time(40*time.Second), 4.0},
							{start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(2*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
				}
			},
		},
		{
			name:     "nonoverlapping_aligned",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
			// Use a time that is aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, 0.0},
							{start + execute.Time(10*time.Second), 1.0},
							{start + execute.Time(20*time.Second), 2.0},
							{start + execute.Time(30*time.Second), 3.0},
							{start + execute.Time(40*time.Second), 4.0},
							{start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(2*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
				}
			},
		},
		{
			name:     "overlapping_nonaligned",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
			// Use a time that is *not* aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(2 * time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, 0.0},
							{start + execute.Time(10*time.Second), 1.0},
							{start + execute.Time(20*time.Second), 2.0},
							{start + execute.Time(30*time.Second), 3.0},
							{start + execute.Time(40*time.Second), 4.0},
							{start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(2*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, 0.0},
							{start + execute.Time(10*time.Second), 1.0},
							{start + execute.Time(20*time.Second), 2.0},
							{start + execute.Time(30*time.Second), 3.0},
							{start + execute.Time(40*time.Second), 4.0},
							{start + execute.Time(50*time.Second), 5.0},
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
				}
			},
		},
		{
			name:     "overlapping_aligned",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
			// Use a time that is aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(2 * time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, 0.0},
							{start + execute.Time(10*time.Second), 1.0},
							{start + execute.Time(20*time.Second), 2.0},
							{start + execute.Time(30*time.Second), 3.0},
							{start + execute.Time(40*time.Second), 4.0},
							{start + execute.Time(50*time.Second), 5.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(2*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, 0.0},
							{start + execute.Time(10*time.Second), 1.0},
							{start + execute.Time(20*time.Second), 2.0},
							{start + execute.Time(30*time.Second), 3.0},
							{start + execute.Time(40*time.Second), 4.0},
							{start + execute.Time(50*time.Second), 5.0},
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(120*time.Second), 12.0},
							{start + execute.Time(130*time.Second), 13.0},
							{start + execute.Time(140*time.Second), 14.0},
						},
					},
				}
			},
		},
		{
			name:     "underlapping_nonaligned",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
			// Use a time that is *not* aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 10, 10, 10, time.UTC).UnixNano()),
			every:  execute.Duration(2 * time.Minute),
			period: execute.Duration(time.Minute),
			num:    24,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start + 1*execute.Time(time.Minute),
							Stop:  start + 2*execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(3*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(180*time.Second), 18.0},
							{start + execute.Time(190*time.Second), 19.0},
							{start + execute.Time(200*time.Second), 20.0},
							{start + execute.Time(210*time.Second), 21.0},
							{start + execute.Time(220*time.Second), 22.0},
							{start + execute.Time(230*time.Second), 23.0},
						},
					},
				}
			},
		},
		{
			name:     "underlapping_aligned",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
			// Use a time that is  aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(2 * time.Minute),
			period: execute.Duration(time.Minute),
			num:    24,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start + 1*execute.Time(time.Minute),
							Stop:  start + 2*execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), 6.0},
							{start + execute.Time(70*time.Second), 7.0},
							{start + execute.Time(80*time.Second), 8.0},
							{start + execute.Time(90*time.Second), 9.0},
							{start + execute.Time(100*time.Second), 10.0},
							{start + execute.Time(110*time.Second), 11.0},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(3*time.Minute),
							Stop:  start + execute.Time(4*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TFloat, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(180*time.Second), 18.0},
							{start + execute.Time(190*time.Second), 19.0},
							{start + execute.Time(200*time.Second), 20.0},
							{start + execute.Time(210*time.Second), 21.0},
							{start + execute.Time(220*time.Second), 22.0},
							{start + execute.Time(230*time.Second), 23.0},
						},
					},
				}
			},
		},
		{
			name:     "nonoverlapping_aligned_int",
			valueCol: execute.ColMeta{Label: "_value", Type: execute.TInt, Kind: execute.ValueColKind},
			// Use a time that is aligned with the every/period durations of the window
			start:  execute.Time(time.Date(2017, 10, 10, 10, 0, 0, 0, time.UTC).UnixNano()),
			every:  execute.Duration(time.Minute),
			period: execute.Duration(time.Minute),
			num:    15,
			want: func(start execute.Time) []*executetest.Block {
				return []*executetest.Block{
					{
						Bnds: execute.Bounds{
							Start: start,
							Stop:  start + execute.Time(time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TInt, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start, int64(0.0)},
							{start + execute.Time(10*time.Second), int64(1)},
							{start + execute.Time(20*time.Second), int64(2)},
							{start + execute.Time(30*time.Second), int64(3)},
							{start + execute.Time(40*time.Second), int64(4)},
							{start + execute.Time(50*time.Second), int64(5)},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(1*time.Minute),
							Stop:  start + execute.Time(2*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TInt, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(60*time.Second), int64(6)},
							{start + execute.Time(70*time.Second), int64(7)},
							{start + execute.Time(80*time.Second), int64(8)},
							{start + execute.Time(90*time.Second), int64(9)},
							{start + execute.Time(100*time.Second), int64(10)},
							{start + execute.Time(110*time.Second), int64(11)},
						},
					},
					{
						Bnds: execute.Bounds{
							Start: start + execute.Time(2*time.Minute),
							Stop:  start + execute.Time(3*time.Minute),
						},
						ColMeta: []execute.ColMeta{
							{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
							{Label: "_value", Type: execute.TInt, Kind: execute.ValueColKind},
						},
						Data: [][]interface{}{
							{start + execute.Time(120*time.Second), int64(12)},
							{start + execute.Time(130*time.Second), int64(13)},
							{start + execute.Time(140*time.Second), int64(14)},
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			start := tc.start
			stop := start + execute.Time(time.Hour)

			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewBlockBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(execute.DefaultTriggerSpec)

			fw := functions.NewFixedWindowTransformation(
				d,
				c,
				execute.Bounds{
					Start: start,
					Stop:  stop,
				},
				execute.Window{
					Every:  tc.every,
					Period: tc.period,
					Start:  start,
				},
			)

			block0 := &executetest.Block{
				Bnds: execute.Bounds{
					Start: start,
					Stop:  stop,
				},
				ColMeta: []execute.ColMeta{
					{Label: "_time", Type: execute.TTime, Kind: execute.TimeColKind},
					tc.valueCol,
				},
			}

			for i := 0; i < tc.num; i++ {
				var v interface{}
				switch tc.valueCol.Type {
				case execute.TBool:
					v = bool(i%2 == 0)
				case execute.TInt:
					v = int64(i)
				case execute.TUInt:
					v = uint64(i)
				case execute.TFloat:
					v = float64(i)
				case execute.TString:
					v = strconv.Itoa(i)
				}
				block0.Data = append(block0.Data, []interface{}{
					start + execute.Time(time.Duration(i)*10*time.Second),
					v,
				})
			}

			parentID := executetest.RandomDatasetID()
			if err := fw.Process(parentID, block0); err != nil {
				t.Fatal(err)
			}

			got := executetest.BlocksFromCache(c)

			sort.Sort(executetest.SortedBlocks(got))
			want := tc.want(start)
			sort.Sort(executetest.SortedBlocks(want))

			if !cmp.Equal(want, got) {
				t.Errorf("unexpected blocks -want/+got\n%s", cmp.Diff(want, got))
			}
		})
	}
}
