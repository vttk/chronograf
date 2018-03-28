package functions_test

import (
	"math/rand"
	"time"

	"github.com/gonum/stat/distuv"
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/execute/executetest"
)

const (
	N     = 1e6
	Mu    = 10
	Sigma = 3

	seed = 42
)

func init() {
	query.FinalizeRegistration()
}

// NormalData is a slice of N random values that are normaly distributed with mean Mu and standard deviation Sigma.
var NormalData []float64

// NormalBlock is a block of data whose value col is NormalData.
var NormalBlock execute.Block

func init() {
	dist := distuv.Normal{
		Mu:     Mu,
		Sigma:  Sigma,
		Source: rand.New(rand.NewSource(seed)),
	}
	NormalData = make([]float64, N)
	for i := range NormalData {
		NormalData[i] = dist.Rand()
	}
	normalBlockBuilder := execute.NewColListBlockBuilder(executetest.UnlimitedAllocator)
	normalBlockBuilder.SetBounds(execute.Bounds{
		Start: execute.Time(time.Date(2016, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
		Stop:  execute.Time(time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano()),
	})

	normalBlockBuilder.AddCol(execute.TimeCol)
	normalBlockBuilder.AddCol(execute.ColMeta{Label: execute.DefaultValueColLabel, Type: execute.TFloat, Kind: execute.ValueColKind})
	normalBlockBuilder.AddCol(execute.ColMeta{Label: "t1", Type: execute.TString, Kind: execute.TagColKind, Common: true})
	normalBlockBuilder.AddCol(execute.ColMeta{Label: "t2", Type: execute.TString, Kind: execute.TagColKind, Common: false})

	times := make([]execute.Time, N)
	values := NormalData
	t1 := "a"
	t2 := make([]string, N)

	start := normalBlockBuilder.Bounds().Start
	for i, v := range values {
		// There are roughly 1 million, 31 second intervals in a year.
		times[i] = start + execute.Time(time.Duration(i*31)*time.Second)
		// Pick t2 based off the value
		switch int(v) % 3 {
		case 0:
			t2[i] = "x"
		case 1:
			t2[i] = "y"
		case 2:
			t2[i] = "z"
		}
	}

	normalBlockBuilder.AppendTimes(0, times)
	normalBlockBuilder.AppendFloats(1, values)
	normalBlockBuilder.SetCommonString(2, t1)
	normalBlockBuilder.AppendStrings(3, t2)

	NormalBlock, _ = normalBlockBuilder.Block()
}
