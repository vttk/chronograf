package functions

import (
	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/execute"
	"github.com/influxdata/ifql/query/plan"
)

const SumKind = "sum"

type SumOpSpec struct {
}

var sumSignature = query.DefaultFunctionSignature()

func init() {
	query.RegisterFunction(SumKind, createSumOpSpec, sumSignature)
	query.RegisterOpSpec(SumKind, newSumOp)
	plan.RegisterProcedureSpec(SumKind, newSumProcedure, SumKind)
	execute.RegisterTransformation(SumKind, createSumTransformation)
}

func createSumOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	return new(SumOpSpec), nil
}

func newSumOp() query.OperationSpec {
	return new(SumOpSpec)
}

func (s *SumOpSpec) Kind() query.OperationKind {
	return SumKind
}

type SumProcedureSpec struct {
}

func newSumProcedure(query.OperationSpec, plan.Administration) (plan.ProcedureSpec, error) {
	return new(SumProcedureSpec), nil
}

func (s *SumProcedureSpec) Kind() plan.ProcedureKind {
	return SumKind
}

func (s *SumProcedureSpec) Copy() plan.ProcedureSpec {
	return new(SumProcedureSpec)
}

func (s *SumProcedureSpec) AggregateMethod() string {
	return SumKind
}
func (s *SumProcedureSpec) ReAggregateSpec() plan.ProcedureSpec {
	return new(SumProcedureSpec)
}

func (s *SumProcedureSpec) PushDownRules() []plan.PushDownRule {
	return []plan.PushDownRule{{
		Root:    FromKind,
		Through: nil,
		Match: func(spec plan.ProcedureSpec) bool {
			selectSpec := spec.(*FromProcedureSpec)
			return !selectSpec.GroupingSet
		},
	}}
}
func (s *SumProcedureSpec) PushDown(root *plan.Procedure, dup func() *plan.Procedure) {
	selectSpec := root.Spec.(*FromProcedureSpec)
	if selectSpec.AggregateSet {
		root = dup()
		selectSpec = root.Spec.(*FromProcedureSpec)
		selectSpec.AggregateSet = false
		selectSpec.AggregateMethod = ""
		return
	}
	selectSpec.AggregateSet = true
	selectSpec.AggregateMethod = s.AggregateMethod()
}

type SumAgg struct{}

func createSumTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, a.Bounds(), new(SumAgg), a.Allocator())
	return t, d, nil
}
func (a *SumAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}
func (a *SumAgg) NewIntAgg() execute.DoIntAgg {
	return new(SumIntAgg)
}
func (a *SumAgg) NewUIntAgg() execute.DoUIntAgg {
	return new(SumUIntAgg)
}
func (a *SumAgg) NewFloatAgg() execute.DoFloatAgg {
	return new(SumFloatAgg)
}
func (a *SumAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

type SumIntAgg struct {
	sum int64
}

func (a *SumIntAgg) DoInt(vs []int64) {
	for _, v := range vs {
		a.sum += v
	}
}
func (a *SumIntAgg) Type() execute.DataType {
	return execute.TInt
}
func (a *SumIntAgg) ValueInt() int64 {
	return a.sum
}

type SumUIntAgg struct {
	sum uint64
}

func (a *SumUIntAgg) DoUInt(vs []uint64) {
	for _, v := range vs {
		a.sum += v
	}
}
func (a *SumUIntAgg) Type() execute.DataType {
	return execute.TUInt
}
func (a *SumUIntAgg) ValueUInt() uint64 {
	return a.sum
}

type SumFloatAgg struct {
	sum float64
}

func (a *SumFloatAgg) DoFloat(vs []float64) {
	for _, v := range vs {
		a.sum += v
	}
}
func (a *SumFloatAgg) Type() execute.DataType {
	return execute.TFloat
}
func (a *SumFloatAgg) ValueFloat() float64 {
	return a.sum
}
