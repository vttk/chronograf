package functions

import (
	"fmt"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
	"github.com/influxdata/ifql/semantic"
)

const LastKind = "last"

type LastOpSpec struct {
	Column     string `json:"column"`
	UseRowTime bool   `json:"useRowtime"`
}

var lastSignature = query.DefaultFunctionSignature()

func init() {
	lastSignature.Params["column"] = semantic.String
	lastSignature.Params["useRowTime"] = semantic.Bool

	query.RegisterFunction(LastKind, createLastOpSpec, lastSignature)
	query.RegisterOpSpec(LastKind, newLastOp)
	plan.RegisterProcedureSpec(LastKind, newLastProcedure, LastKind)
	execute.RegisterTransformation(LastKind, createLastTransformation)
}

func createLastOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(LastOpSpec)
	if c, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = c
	}
	if useRowTime, ok, err := args.GetBool("useRowTime"); err != nil {
		return nil, err
	} else if ok {
		spec.UseRowTime = useRowTime
	}
	return spec, nil
}

func newLastOp() query.OperationSpec {
	return new(LastOpSpec)
}

func (s *LastOpSpec) Kind() query.OperationKind {
	return LastKind
}

type LastProcedureSpec struct {
	Column     string
	UseRowTime bool
}

func newLastProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LastOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &LastProcedureSpec{
		Column:     spec.Column,
		UseRowTime: spec.UseRowTime,
	}, nil
}

func (s *LastProcedureSpec) Kind() plan.ProcedureKind {
	return LastKind
}

func (s *LastProcedureSpec) PushDownRules() []plan.PushDownRule {
	return []plan.PushDownRule{{
		Root:    FromKind,
		Through: []plan.ProcedureKind{GroupKind, LimitKind, FilterKind},
		Match: func(spec plan.ProcedureSpec) bool {
			selectSpec := spec.(*FromProcedureSpec)
			return !selectSpec.AggregateSet
		},
	}}
}

func (s *LastProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*FromProcedureSpec)
	if selectSpec.BoundsSet || selectSpec.LimitSet || selectSpec.DescendingSet {
		root = dup()
		selectSpec = root.Spec.(*FromProcedureSpec)
		selectSpec.BoundsSet = false
		selectSpec.Bounds = plan.BoundsSpec{}
		selectSpec.LimitSet = false
		selectSpec.PointsLimit = 0
		selectSpec.SeriesLimit = 0
		selectSpec.SeriesOffset = 0
		selectSpec.DescendingSet = false
		selectSpec.Descending = false
		return
	}
	selectSpec.BoundsSet = true
	selectSpec.Bounds = plan.BoundsSpec{
		Start: query.MinTime,
		Stop:  query.Now,
	}
	selectSpec.LimitSet = true
	selectSpec.PointsLimit = 1
	selectSpec.DescendingSet = true
	selectSpec.Descending = true
}

func (s *LastProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LastProcedureSpec)
	ns.Column = s.Column
	ns.UseRowTime = s.UseRowTime
	return ns
}

type LastSelector struct {
	rows []execute.Row
}

func createLastTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*LastProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	t, d := execute.NewRowSelectorTransformationAndDataset(id, mode, a.Bounds(), new(LastSelector), ps.Column, ps.UseRowTime, a.Allocator())
	return t, d, nil
}

func (s *LastSelector) reset() {
	s.rows = nil
}
func (s *LastSelector) NewBoolSelector() execute.DoBoolRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewIntSelector() execute.DoIntRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewUIntSelector() execute.DoUIntRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewFloatSelector() execute.DoFloatRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) NewStringSelector() execute.DoStringRowSelector {
	s.reset()
	return s
}

func (s *LastSelector) Rows() []execute.Row {
	return s.rows
}

func (s *LastSelector) selectLast(l int, rr execute.RowReader) {
	if l > 0 {
		s.rows = []execute.Row{execute.ReadRow(l-1, rr)}
	}
}

func (s *LastSelector) DoBool(vs []bool, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoInt(vs []int64, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoUInt(vs []uint64, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoFloat(vs []float64, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
func (s *LastSelector) DoString(vs []string, rr execute.RowReader) {
	s.selectLast(len(vs), rr)
}
