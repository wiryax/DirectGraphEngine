package engine

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func strPtr(s string) *string {
	return &s
}

type MockTask struct {
	task func(gCtx *GraphContext) error
}

func NewMockTask(task func(gCtx *GraphContext) error) *MockTask {
	return &MockTask{
		task: task,
	}
}

func (m *MockTask) Execute(gCtx *GraphContext) error {
	return m.task(gCtx)
}

func newVertex(id string, task func(gCtx *GraphContext) error) *Vertex {
	mockTask := NewMockTask(task)
	return &Vertex{id: id, task: mockTask, state: Pendding}
}

func newEdge(from, to *Vertex) *Edge {
	return &Edge{from: from, to: to}
}

func newRuntimeState(state map[string]*Vertex) *RuntimeState {
	return &RuntimeState{state: state}
}

func newExpression(tk []token) expression {
	return expression{tokens: tk}
}

var taskFunc = func(err error) *MockTask {
	return &MockTask{
		task: func(gCtx *GraphContext) error {
			return err
		},
	}
}

type (
	tVertex struct {
		id   string
		task Task
	}

	relationship struct {
		from,
		to string
		lOp    tokenType
		pConst state
		tk     []token
	}

	tGraph struct {
		tVertex      []tVertex
		relationship []relationship
		log          GraphLogger
	}
)

func TestGraphWorkflow(t *testing.T) {
	testCase := []struct {
		title  string
		tGraph tGraph
		preRuntimeState,
		postRuntimeState *RuntimeState
		preVertex,
		postVertex []*Vertex
	}{
		{
			title: "TestSingelParentDeps_SuccessCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(nil),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpAnd,
						pConst: Success,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*Vertex),
				variable: make(map[string]string),
				vState:   make(map[*Vertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*Vertex),
				variable: make(map[string]string),
				vState:   make(map[*Vertex]state),
			},
			preVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Pendding,
					penddingEdge: 0,
					failEdge:     0,
				}, &Vertex{
					id:           "vB",
					state:        Pendding,
					penddingEdge: 1,
					failEdge:     0,
				},
			},
			postVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Success,
					penddingEdge: 0,
					failEdge:     0,
				},
				&Vertex{
					id:           "vB",
					state:        Success,
					penddingEdge: 0,
					failEdge:     0,
				},
			},
		}, {
			title: "TestSingelParentDeps_FailCase",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("unexpected error occure while execute task vertex")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpAnd,
						pConst: Success,
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state:    make(map[string]*Vertex),
				variable: make(map[string]string),
				vState:   make(map[*Vertex]state),
			},
			postRuntimeState: &RuntimeState{
				state:    make(map[string]*Vertex),
				variable: make(map[string]string),
				vState:   make(map[*Vertex]state),
			},
			preVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Pendding,
					penddingEdge: 0,
					failEdge:     0,
				}, &Vertex{
					id:           "vB",
					state:        Pendding,
					penddingEdge: 1,
					failEdge:     0,
				},
			},
			postVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Fail,
					penddingEdge: 0,
					failEdge:     0,
				},
				&Vertex{
					id:           "vB",
					state:        Skipped,
					penddingEdge: 0,
					failEdge:     1,
				},
			},
		}, {
			title: "TestSingelParentDeps_FailCase_with_OR_logical_true",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("unexpected error occure while execute task vertex")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpOr,
						pConst: Success,
						tk: []token{
							{
								id:    "isValidDate",
								eType: ExpVariable,
							}, {
								id:    "currentDate",
								eType: ExpVariable,
							}, {
								id:    "",
								eType: ExpEqual,
							},
						},
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state: make(map[string]*Vertex),
				variable: map[string]string{
					"isValidDate": time.Now().Format("02-01-2006"),
					"currentDate": time.Now().Format("02-01-2006"),
				},
				vState: make(map[*Vertex]state),
			},
			postRuntimeState: &RuntimeState{
				state: make(map[string]*Vertex),
				variable: map[string]string{
					"isValidDate": time.Now().Format("02-01-2006"),
					"currentDate": time.Now().Format("02-01-2006"),
				},
				vState: make(map[*Vertex]state),
			},
			preVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Pendding,
					penddingEdge: 0,
					failEdge:     0,
				}, &Vertex{
					id:           "vB",
					state:        Pendding,
					penddingEdge: 1,
					failEdge:     0,
				},
			},
			postVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Fail,
					penddingEdge: 0,
					failEdge:     0,
				},
				&Vertex{
					id:           "vB",
					state:        Success,
					penddingEdge: 0,
					failEdge:     0,
				},
			},
		}, {
			title: "TestSingelParentDeps_FailCase_with_OR_logical_false",
			tGraph: tGraph{
				tVertex: []tVertex{
					{
						id:   "vA",
						task: taskFunc(errors.New("unexpected error occure while execute task vertex")),
					}, {
						id:   "vB",
						task: taskFunc(nil),
					},
				},
				relationship: []relationship{
					{
						from:   "vA",
						to:     "vB",
						lOp:    ExpOr,
						pConst: Success,
						tk: []token{
							{
								id:    "isValidDate",
								eType: ExpVariable,
							}, {
								id:    "currentDate",
								eType: ExpVariable,
							}, {
								id:    "",
								eType: ExpEqual,
							},
						},
					},
				},
				log: nil,
			},
			preRuntimeState: &RuntimeState{
				state: make(map[string]*Vertex),
				variable: map[string]string{
					"isValidDate": "02-06-2026",
					"currentDate": time.Now().Format("01-02-2006"),
				},
				vState: make(map[*Vertex]state),
			},
			postRuntimeState: &RuntimeState{
				state: make(map[string]*Vertex),
				variable: map[string]string{
					"isValidDate": "02-06-2026",
					"currentDate": time.Now().Format("01-02-2006"),
				},
				vState: make(map[*Vertex]state),
			},
			preVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Pendding,
					penddingEdge: 0,
					failEdge:     0,
				}, &Vertex{
					id:           "vB",
					state:        Pendding,
					penddingEdge: 1,
					failEdge:     0,
				},
			},
			postVertex: []*Vertex{
				&Vertex{
					id:           "vA",
					state:        Fail,
					penddingEdge: 0,
					failEdge:     0,
				},
				&Vertex{
					id:           "vB",
					state:        Skipped,
					penddingEdge: 0,
					failEdge:     1,
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.title, func(t *testing.T) {
			g := NewGraph(tc.title)
			gCtx := NewGraphContext(NewLogger(os.Stdout), tc.preRuntimeState)

			for _, v := range tc.tGraph.tVertex {
				g.Add(v.id, v.task)
			}

			for _, r := range tc.tGraph.relationship {
				fV := g.GetVertex(r.from)
				tV := g.GetVertex(r.to)

				if fV == nil || tV == nil {
					t.Fatalf("unexpected while setup test case, cannot find vertex with id %s or %s: from %v, to %v", r.from, r.to, fV, tV)
				}

				g.Connect(fV, tV, r.pConst, r.lOp, r.tk)
			}

			//assert pre test
			if !cmp.Equal(gCtx.rState, tc.preRuntimeState, cmp.AllowUnexported(RuntimeState{}, Vertex{}, Edge{}, expression{}, token{})) {
				t.Errorf("unexpected pre-state test result.\n want\t%+v,\n got\t%+v", tc.preRuntimeState, gCtx.rState)
			}

			if !cmp.Equal(g.vertex, tc.preVertex, cmp.AllowUnexported(Vertex{}), cmpopts.IgnoreFields(Vertex{}, "in", "out", "task")) {
				t.Errorf("unexpected pre-vertex result.\n want\t%+v,\n got\t%+v", tc.preVertex, g.vertex)
			}

			g.RunWithContext(gCtx)

			//assert post test
			if !cmp.Equal(gCtx.rState, tc.postRuntimeState, cmp.AllowUnexported(RuntimeState{}, Vertex{}, Edge{}, expression{}, token{})) {
				t.Errorf("unexpected post-state test result.\n want\t%+v,\n got\t%+v", tc.postRuntimeState, gCtx.rState)
			}

			if !cmp.Equal(g.vertex, tc.postVertex, cmp.AllowUnexported(Vertex{}), cmpopts.IgnoreFields(Vertex{}, "in", "out", "task")) {
				t.Errorf("unexpected post-vertex result.\n want\t%+v,\n got\t%+v", tc.postVertex, g.vertex)
			}
		})
	}

}
