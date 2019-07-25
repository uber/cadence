// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package xdc

import (
	"github.com/uber/cadence/.gen/go/shared"
	"math/rand"
	"strings"
	"time"
)

type (
	// Model represents a state transition graph that contains all the relationships of Vertex
	Model interface {
		AddEdge(...Edge)
		ListEdges() []Edge
	}

	// Generator generates a sequence of Vertexes based on the defined models
	// It must define InitialEntryVertex and ExitVertex
	Generator interface {
		// InitialEntryVertex is the beginning vertexes of the graph
		// Only one vertex will be picked as the entry
		AddInitialEntryVertex(...Vertex)
		// ExitVertex is the terminate vertexes of the graph
		AddExitVertex(...Vertex)
		// RandomEntryVertex is a random entry point which can be access at any state of the generator
		AddRandomEntryVertex(...Vertex)
		// AddModel loads model into the generator
		// AddModel can load multiple models and models will be joint if there is common vertexes
		AddModel(Model)
		// HasNextVertex determines if there is more vertex to generate
		HasNextVertex() bool
		// GetNextVertex generates next vertex batch
		GetNextVertex() []Vertex
		// ListGeneratedVertex lists the pasted generated vertexes
		ListGeneratedVertex() []Vertex
		// Reset resets the generator to initial state
		Reset()
	}

	// Vertex represents a state in the model. A state represents a type of an Cadence event
	Vertex interface {
		// The name of the vertex. Usually, this will be the Cadence event type
		SetName(string)
		GetName() string
		Equals(Vertex) bool
		// IsStrictOnNextVertex means if the vertex must be followed by its children
		// When IsStrictOnNextVertex set to true, it means this event can only follow by its neighbors
		SetIsStrictOnNextVertex(bool)
		IsStrictOnNextVertex() bool
		// MaxNextVertex means the max neighbors can branch out from this vertex
		// MaxNextVertex means the max neighbors can branch out from this vertex
		SetMaxNextVertex(int)
		GetMaxNextVertex() int
	}

	// Edge is the connection between two vertexes
	Edge interface {
		// StartVertex is the head of the connection
		SetStartVertex(Vertex)
		GetStartVertex() Vertex
		// EndVertex is the end of the connection
		SetEndVertex(Vertex)
		GetEndVertex() Vertex
		// Condition defines a function to determine if this connection is accessible
		SetCondition(func() bool)
		GetCondition() func() bool
		// Action defines function to perform when the end vertex reached
		SetAction(func())
		GetAction() func()
	}

	// HistoryEvent is the history event vertex
	HistoryEvent struct {
		name                 string
		isStrictOnNextVertex bool
		maxNextGeneration    int
	}

	// HistoryEventModel is the history event model
	HistoryEventModel struct {
		edges []Edge
	}

	// EventGenerator is a event generator
	EventGenerator struct {
		connections         map[Vertex][]Edge
		previousVertexes    []Vertex
		leafVertexes        []Vertex
		entryVertexes       []Vertex
		exitVertexes        map[Vertex]bool
		randomEntryVertexes []Vertex
		dice                *rand.Rand
	}

	// Connection is the edge
	Connection struct {
		startVertex Vertex
		endVertex   Vertex
		condition   func() bool
		action      func()
	}

	// RevokeFunc is the condition inside connection
	RevokeFunc struct {
		methodName string
		input      []interface{}
	}
)

// NewEventGenerator initials the event generator
func NewEventGenerator() Generator {
	return &EventGenerator{
		connections:         make(map[Vertex][]Edge),
		previousVertexes:    make([]Vertex, 0),
		leafVertexes:        make([]Vertex, 0),
		entryVertexes:       make([]Vertex, 0),
		exitVertexes:        make(map[Vertex]bool),
		randomEntryVertexes: make([]Vertex, 0),
		dice:                rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// AddInitialEntryVertex adds the initial entry vertex. Generator will only start from one of the entry vertex
func (g *EventGenerator) AddInitialEntryVertex(entry ...Vertex) {
	g.entryVertexes = append(g.entryVertexes, entry...)
}

// AddExitVertex adds the terminate vertex in the generator
func (g *EventGenerator) AddExitVertex(exit ...Vertex) {
	for _, v := range exit {
		g.exitVertexes[v] = true
	}
}

// AddRandomEntryVertex adds the random vertex in the generator
func (g *EventGenerator) AddRandomEntryVertex(exit ...Vertex) {
	g.randomEntryVertexes = append(g.randomEntryVertexes, exit...)
}

// AddModel adds a model
func (g *EventGenerator) AddModel(model Model) {
	for _, e := range model.ListEdges() {
		if _, ok := g.connections[e.GetStartVertex()]; !ok {
			g.connections[e.GetStartVertex()] = make([]Edge, 0)
		}
		g.connections[e.GetStartVertex()] = append(g.connections[e.GetStartVertex()], e)
	}
}

// ListGeneratedVertex returns all the generated vertex
func (g EventGenerator) ListGeneratedVertex() []Vertex {
	return g.previousVertexes
}

// HasNextVertex checks if there is accessible vertex
func (g *EventGenerator) HasNextVertex() bool {
	for _, prev := range g.previousVertexes {
		if _, ok := g.exitVertexes[prev]; ok {
			return false
		}
	}
	return len(g.leafVertexes) > 0 || (len(g.previousVertexes) == 0 && len(g.entryVertexes) > 0)
}

// GetNextVertex generates the next vertex
func (g *EventGenerator) GetNextVertex() []Vertex {
	if !g.HasNextVertex() {
		panic("Generator reached to a terminate state.")
	}

	batch := make([]Vertex, 0)
	for g.HasNextVertex() && g.canBatch(batch) {
		res := make([]Vertex, 0)
		switch {
		case len(g.previousVertexes) == 0:
			// Generate for the first time, get the event candidates from entry vertex group
			res = append(res, g.getEntryVertex())
		case len(g.randomEntryVertexes) > 0 && g.dice.Intn(len(g.connections)) == 0:
			// Get the event candidate from random vertex group
			res = append(res, g.getRandomVertex())
		default:
			// Get the event candidates based on context
			idx := g.getVertexIndexFromLeaf()
			res = append(res, g.randomNextVertex(idx)...)
			g.leafVertexes = append(g.leafVertexes[:idx], g.leafVertexes[idx+1:]...)
		}
		g.leafVertexes = append(g.leafVertexes, res...)
		g.previousVertexes = append(g.previousVertexes, res...)
		batch = append(batch, res...)
	}
	return batch
}

// Reset reset the generator to the initial state
func (g *EventGenerator) Reset() {
	g.previousVertexes = make([]Vertex, 0)
	g.leafVertexes = make([]Vertex, 0)
	g.dice = rand.New(rand.NewSource(time.Now().Unix()))
}

func (g *EventGenerator) canBatch(history []Vertex) bool {
	if len(history) == 0 {
		return true
	}

	hasPendingDecisionTask := false
	for _, event := range g.previousVertexes {
		switch event.GetName() {
		case shared.EventTypeDecisionTaskScheduled.String():
			hasPendingDecisionTask = true
		case shared.EventTypeDecisionTaskCompleted.String(),
			shared.EventTypeDecisionTaskFailed.String(),
			shared.EventTypeDecisionTaskTimedOut.String():
			hasPendingDecisionTask = false
		}
	}
	if hasPendingDecisionTask {
		return false
	}
	if history[len(history)-1].GetName() == shared.EventTypeDecisionTaskScheduled.String() {
		return false
	}
	if history[0].GetName() == shared.EventTypeDecisionTaskCompleted.String() {
		return len(history) == 1
	}
	return true
}

func (g *EventGenerator) getEntryVertex() Vertex {
	if len(g.entryVertexes) == 0 {
		panic("No possible start vertex to go to next step")
	}
	nextRange := len(g.entryVertexes)
	nextIdx := g.dice.Intn(nextRange)
	vertex := g.entryVertexes[nextIdx]
	return vertex
}

func (g *EventGenerator) getRandomVertex() Vertex {
	if len(g.randomEntryVertexes) == 0 {
		panic("No possible vertex to go to next step")
	}
	nextRange := len(g.randomEntryVertexes)
	nextIdx := g.dice.Intn(nextRange)
	vertex := g.randomEntryVertexes[nextIdx]
	return vertex
}

func (g *EventGenerator) getVertexIndexFromLeaf() int {
	if len(g.leafVertexes) == 0 {
		panic("No possible vertex to go to next step")
	}

	isAccessible := false
	nextRange := len(g.leafVertexes)
	notAvailable := make(map[int]bool)
	var leaf Vertex
	var nextVertexIdx int
	for !isAccessible {
		nextVertexIdx = g.dice.Intn(nextRange)
		if _, ok := notAvailable[nextVertexIdx]; ok {
			continue
		}
		leaf = g.leafVertexes[nextVertexIdx]
		if g.leafVertexes[len(g.leafVertexes)-1].IsStrictOnNextVertex() {
			nextVertexIdx = len(g.leafVertexes) - 1
			leaf = g.leafVertexes[nextVertexIdx]
		}
		neighbors := g.connections[leaf]
		for _, nextV := range neighbors {
			if nextV.GetCondition() == nil || (nextV.GetCondition() != nil && nextV.GetCondition()()) {
				isAccessible = true
				break
			}
		}
		if !isAccessible {
			notAvailable[nextVertexIdx] = true
			if len(notAvailable) == nextRange {
				panic("cannot find vertex to proceed")
			}
		}
	}
	return nextVertexIdx
}

func (g *EventGenerator) randomNextVertex(nextVertexIdx int) []Vertex {
	nextVertex := g.leafVertexes[nextVertexIdx]
	count := g.dice.Intn(nextVertex.GetMaxNextVertex()) + 1
	neighbors := g.connections[nextVertex]
	neighborsRange := len(neighbors)
	res := make([]Vertex, 0)
	for i := 0; i < count; i++ {
		nextIdx := g.dice.Intn(neighborsRange)
		for neighbors[nextIdx].GetCondition() != nil && !neighbors[nextIdx].GetCondition()() {
			nextIdx = g.dice.Intn(neighborsRange)
		}
		newConnection := neighbors[nextIdx]
		newLeaf := newConnection.GetEndVertex()
		res = append(res, newLeaf)
		if newConnection.GetAction() != nil {
			newConnection.GetAction()()
		}
		if _, ok := g.exitVertexes[newLeaf]; ok {
			res = []Vertex{newLeaf}
			return res
		}
	}

	return res
}

// NewConnection initials a new connection
func NewConnection(
	start Vertex,
	end Vertex,
) Edge {
	return &Connection{
		startVertex: start,
		endVertex:   end,
	}
}

// SetStartVertex sets the start vertex
func (c *Connection) SetStartVertex(start Vertex) {
	c.startVertex = start
}

// GetStartVertex returns the start vertex
func (c Connection) GetStartVertex() Vertex {
	return c.startVertex
}

// SetEndVertex sets the end vertex
func (c *Connection) SetEndVertex(end Vertex) {
	c.endVertex = end
}

// GetEndVertex returns the end vertex
func (c Connection) GetEndVertex() Vertex {
	return c.endVertex
}

// SetCondition sets the condition to access this edge
func (c *Connection) SetCondition(condition func() bool) {
	c.condition = condition
}

// GetCondition returns the condition
func (c Connection) GetCondition() func() bool {
	return c.condition
}

// SetAction sets an action to perform when the end vertex hits
func (c Connection) SetAction(action func()) {
	c.action = action
}

// GetAction returns the action
func (c Connection) GetAction() func() {
	return c.action
}

// NewHistoryEvent initials a history event
func NewHistoryEvent(name string) Vertex {
	return &HistoryEvent{
		name:                 name,
		isStrictOnNextVertex: false,
		maxNextGeneration:    1,
	}
}

// GetName returns the name
func (he HistoryEvent) GetName() string {
	return he.name
}

// SetName sets the name
func (he *HistoryEvent) SetName(name string) {
	he.name = name
}

// Equals compares two vertex
func (he *HistoryEvent) Equals(v Vertex) bool {
	return strings.EqualFold(he.name, v.GetName())
}

// SetIsStrictOnNextVertex sets if a vertex can be added between the current vertex and its child vertexes
func (he *HistoryEvent) SetIsStrictOnNextVertex(isStrict bool) {
	he.isStrictOnNextVertex = isStrict
}

// IsStrictOnNextVertex returns the isStrict flag
func (he HistoryEvent) IsStrictOnNextVertex() bool {
	return he.isStrictOnNextVertex
}

// SetMaxNextVertex sets the max concurrent path can be generated from this vertex
func (he *HistoryEvent) SetMaxNextVertex(maxNextGeneration int) {
	if maxNextGeneration < 1 {
		panic("max next vertex number cannot less than 1")
	}
	he.maxNextGeneration = maxNextGeneration
}

// GetMaxNextVertex returns the max concurrent path
func (he HistoryEvent) GetMaxNextVertex() int {
	return he.maxNextGeneration
}

// NewHistoryEventModel initials new history event model
func NewHistoryEventModel() Model {
	return &HistoryEventModel{
		edges: make([]Edge, 0),
	}
}

// AddEdge adds an edge to the model
func (m *HistoryEventModel) AddEdge(edge ...Edge) {
	m.edges = append(m.edges, edge...)
}

// ListEdges returns all added edges
func (m HistoryEventModel) ListEdges() []Edge {
	return m.edges
}
