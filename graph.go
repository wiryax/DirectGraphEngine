package DirectGraphEngine

import (
	"fmt"
)

type state int

const (
	Pending state = iota
	Success
	Fail
	Running
	Skipped
)

type Edge struct {
	from, to *Vertex
	exp      expression
	lOp      tokenType
	pConst   state
}

func (e *Edge) evalConst() bool {
	return e.from.state == e.pConst
}

type Task interface {
	Execute(gCtx *GraphContext) error
}

type Vertex struct {
	id                    string
	state                 state
	task                  Task
	pendingEdge, failEdge int
	in, out               []*Edge
}

func (v *Vertex) String() string {
	return fmt.Sprintf("id=%s state=%d pendingEdge=%d failEdge=%d,", v.id, v.state, v.pendingEdge, v.failEdge)
}

func (v *Vertex) GetId() string {
	return v.id
}

func (v *Vertex) SetState(s state) {
	v.state = s
}

func (v *Vertex) GetState() state {
	return v.state
}

func (v *Vertex) ExecuteTask(gCtx *GraphContext) error {
	return v.task.Execute(gCtx)
}

type Graph struct {
	id     string
	vertex []*Vertex
}

func NewGraph(id string) *Graph {
	return &Graph{
		id: id,
	}
}

func (g *Graph) RunWithContext(gCtx *GraphContext) {
	g.run(gCtx)
}

func (g *Graph) Run() {
	rState := &RuntimeState{
		variable: make(map[string]string),
	}

	gCtx := NewGraphContext(NewLogger(nil), rState)

	g.run(gCtx)
}

func (g *Graph) run(gCtx *GraphContext) {
	if gCtx == nil {
		panic("graph context cannot nil")
	}

	var queue []*Vertex
	for _, v := range g.vertex {
		if v.pendingEdge == 0 && v.failEdge == 0 {
			queue = append(queue, v)
		}
	}

	g.execute(gCtx, queue)
}

func (g *Graph) execute(gCtx *GraphContext, queue []*Vertex) {
	for {
		if len(queue) == 0 {
			break
		}

		v := queue[0]
		queue = queue[1:]

		v.state = Running

		gCtx.Log(EventStart, LevelInfo, "Start execute vertex", v.GetId(), g.id)
		err := v.ExecuteTask(gCtx)
		if err != nil {
			gCtx.Log(EventFailed, LevelInfo, err.Error(), v.GetId(), g.id)
			v.state = Fail
		} else {
			gCtx.Log(EventSuccess, LevelInfo, "Finish execute vertex", v.GetId(), g.id)
			v.state = Success
		}
		g.getReadyVertex(gCtx, v, &queue)
	}
}

func (g *Graph) getReadyVertex(gCtx *GraphContext, v *Vertex, queue *[]*Vertex) {
	for _, child := range v.out {
		if !child.evalConst() && child.lOp == ExpAnd {
			child.to.failEdge++
			child.to.pendingEdge--
		} else {
			rEvaluate := evaluate(gCtx, child.exp.tokens)
			if rEvaluate == False {
				child.to.failEdge++
				child.to.pendingEdge--
			} else if rEvaluate == True {
				child.to.pendingEdge--
			}
		}

		if child.to.pendingEdge == 0 && child.to.failEdge > 0 {
			child.to.state = Skipped
		}

		g.getReadyVertex(gCtx, child.to, queue)

		if child.to.pendingEdge == 0 && child.to.failEdge == 0 {
			*queue = append(*queue, child.to)
		}
	}
}

func (g *Graph) AddVertex(vertex ...*Vertex) {
	g.vertex = append(g.vertex, vertex...)
}

func (g *Graph) Connect(from, to *Vertex, op state, lOp tokenType, tk []token) {
	edge := &Edge{
		from:   from,
		to:     to,
		pConst: op,
		lOp:    lOp,
	}

	to.pendingEdge++

	for i := range tk {
		edge.exp.push(tk[i])
	}

	from.out = append(from.out, edge)
}

func (g *Graph) Add(id string, task Task) *Vertex {
	v := &Vertex{
		id:    id,
		task:  task,
		state: Pending,
	}

	g.vertex = append(g.vertex, v)
	return v
}

func (g *Graph) GetVertex(id string) *Vertex {
	for i := range g.vertex {
		if g.vertex[i].id == id {
			return g.vertex[i]
		}
	}
	return nil
}
