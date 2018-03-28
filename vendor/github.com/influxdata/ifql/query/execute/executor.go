package execute

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/influxdata/ifql/query"
	"github.com/influxdata/ifql/query/plan"
	"github.com/pkg/errors"
)

type Executor interface {
	Execute(context.Context, *plan.PlanSpec) (map[string]Result, error)
}

type executor struct {
	c Config
}

type Config struct {
	StorageReader StorageReader
}

func NewExecutor(c Config) Executor {
	e := &executor{
		c: c,
	}
	return e
}

type executionState struct {
	p *plan.PlanSpec
	c *Config

	alloc *Allocator

	resources query.ResourceManagement

	bounds Bounds

	results map[string]Result
	sources []Source

	transports []Transport

	dispatcher *poolDispatcher
}

func (e *executor) Execute(ctx context.Context, p *plan.PlanSpec) (map[string]Result, error) {
	es, err := e.createExecutionState(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize execute state")
	}
	es.do(ctx)
	return es.results, nil
}

func validatePlan(p *plan.PlanSpec) error {
	if p.Resources.ConcurrencyQuota == 0 {
		return errors.New("plan must have a non-zero concurrency quota")
	}
	return nil
}

func (e *executor) createExecutionState(ctx context.Context, p *plan.PlanSpec) (*executionState, error) {
	if err := validatePlan(p); err != nil {
		return nil, errors.Wrap(err, "invalid plan")
	}
	es := &executionState{
		p: p,
		c: &e.c,
		alloc: &Allocator{
			Limit: p.Resources.MemoryBytesQuota,
		},
		resources: p.Resources,
		results:   make(map[string]Result, len(p.Results)),
		// TODO(nathanielc): Have the planner specify the dispatcher throughput
		dispatcher: newPoolDispatcher(10),
		bounds: Bounds{
			Start: Time(p.Bounds.Start.Time(p.Now).UnixNano()),
			Stop:  Time(p.Bounds.Stop.Time(p.Now).UnixNano()),
		},
	}
	for name, yield := range p.Results {
		ds, err := es.createNode(ctx, p.Procedures[yield.ID])
		if err != nil {
			return nil, err
		}
		rs := newResultSink(yield)
		ds.AddTransformation(rs)
		es.results[name] = rs
	}
	return es, nil
}

// DefaultTriggerSpec defines the triggering that should be used for datasets
// whose parent transformation is not a windowing transformation.
var DefaultTriggerSpec = query.AfterWatermarkTriggerSpec{}

type triggeringSpec interface {
	TriggerSpec() query.TriggerSpec
}

func (es *executionState) createNode(ctx context.Context, pr *plan.Procedure) (Node, error) {
	// Build execution context
	ec := executionContext{
		es: es,
	}
	if len(pr.Parents) > 0 {
		ec.parents = make([]DatasetID, len(pr.Parents))
		for i, parentID := range pr.Parents {
			ec.parents[i] = DatasetID(parentID)
		}
	}

	// If source create source
	if createS, ok := procedureToSource[pr.Spec.Kind()]; ok {
		s := createS(pr.Spec, DatasetID(pr.ID), es.c.StorageReader, ec)
		es.sources = append(es.sources, s)
		return s, nil
	}

	createT, ok := procedureToTransformation[pr.Spec.Kind()]
	if !ok {
		return nil, fmt.Errorf("unsupported procedure %v", pr.Spec.Kind())
	}

	// Create the transformation
	t, ds, err := createT(DatasetID(pr.ID), AccumulatingMode, pr.Spec, ec)
	if err != nil {
		return nil, err
	}

	// Setup triggering
	var ts query.TriggerSpec = DefaultTriggerSpec
	if t, ok := pr.Spec.(triggeringSpec); ok {
		ts = t.TriggerSpec()
	}
	ds.SetTriggerSpec(ts)

	// Recurse creating parents
	for _, parentID := range pr.Parents {
		parent, err := es.createNode(ctx, es.p.Procedures[parentID])
		if err != nil {
			return nil, err
		}
		transport := newConescutiveTransport(es.dispatcher, t)
		es.transports = append(es.transports, transport)
		parent.AddTransformation(transport)
	}

	return ds, nil
}

func (es *executionState) abort(err error) {
	for _, r := range es.results {
		r.abort(err)
	}
}

func (es *executionState) do(ctx context.Context) {
	for _, src := range es.sources {
		go func(src Source) {
			// Setup panic handling on the source goroutines
			defer func() {
				if e := recover(); e != nil {
					// We had a panic, abort the entire execution.
					var err error
					switch e := e.(type) {
					case error:
						err = e
					default:
						err = fmt.Errorf("%v", e)
					}
					es.abort(fmt.Errorf("panic: %v\n%s", err, debug.Stack()))
				}
			}()
			src.Run(ctx)
		}(src)
	}
	es.dispatcher.Start(es.resources.ConcurrencyQuota, ctx)
	go func() {
		// Wait for all transports to finish
		for _, t := range es.transports {
			select {
			case <-t.Finished():
			case <-ctx.Done():
				es.abort(errors.New("context done"))
			case err := <-es.dispatcher.Err():
				if err != nil {
					es.abort(err)
				}
			}
		}
		// Check for any errors on the dispatcher
		err := es.dispatcher.Stop()
		if err != nil {
			es.abort(err)
		}
	}()
}

type executionContext struct {
	es      *executionState
	parents []DatasetID
}

// Satisfy the ExecutionContext interface

func (ec executionContext) ResolveTime(qt query.Time) Time {
	return Time(qt.Time(ec.es.p.Now).UnixNano())
}
func (ec executionContext) Bounds() Bounds {
	return ec.es.bounds
}

func (ec executionContext) Allocator() *Allocator {
	return ec.es.alloc
}

func (ec executionContext) Parents() []DatasetID {
	return ec.parents
}
func (ec executionContext) ConvertID(id plan.ProcedureID) DatasetID {
	return DatasetID(id)
}
