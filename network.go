package neat

import (
	"log"
	"math/rand"
	"sync/atomic"
)

// The global organism configuration
var config NeatConfig

// Set the global organism configuration
func SetNeatConfig(neatConfig NeatConfig) {
	config = neatConfig
}

// The signature of an activation function
type ActivationFunction func(float64) float64

// Expose the random function so that it can be manipulated by tests
var RandFloat64 = rand.Float64

// Global innovation counter
var innovationCount uint64

// Global identifier counter
var idCount uint64

func nextInnovation() uint64 {
	return atomic.AddUint64(&innovationCount, 1)
}

func nextID() uint64 {
	return atomic.AddUint64(&idCount, 1)
}

// The most general things that can be said about the genes
type gene interface {
	getInnovation() uint64
}

type synapseID uint64
type neuronID uint64
type neuronKind int

// A synapse, the connection between two neurons
type synapse struct {
	id synapseID
	// Sending neuron
	in neuronID
	// Receiving neuron
	out neuronID
	// Weight to apply to signal passing through
	weight float64
	// Synapse activity toggle
	enabled bool
	// Innovation number
	innovation uint64
}

// Create a new synapse from the in neuron to the out neuron
func newSynapse(in, out *neuron) *synapse {
	return &synapse{
		id: synapseID(nextID()),
		in: in.id,
		out: out.id,
		weight: 1.0,
		enabled: true,
		innovation: nextInnovation(),
	}
}

func (s *synapse) clone() *synapse {
	copySynapse := *s
	return &copySynapse
}

func (s *synapse) getInnovation() uint64 {
	return s.innovation
}

// Toggle the enabled state of the synapse
func (s *synapse) toggleEnabled() {
	s.enabled = !s.enabled
}

// Perturbe the weight of a synapse
func (s *synapse) mutateWeight() {
	s.weight = 2 * ((RandFloat64() - 0.5) * config.OrganismConfig.SynapseWeightBound)
}

// The different kinds of neurons
const (
	// Accepts sensory input to an organism
	sensorNeuron neuronKind = iota
	// Expresses the output of an organism
	outputNeuron
	// The "memory" of the organism
	hiddenNeuron
)

// A neuron, a sub-state within the organism. Accepts input from and produces
// output to other neurons.
type neuron struct {
	// Unique id of the neuron
	id neuronID

	// Gene things
	// Current output value
	value float64
	// Innovation number
	innovation uint64
	// Neuron kind
	kind neuronKind

	// Topology things
	// Future output accumulator, if the network is recurrent
	future float64
	// Input sum
	sum float64
	// Visited indicator used to avoid recursion when proagating
	visited bool
	// Seen indicator used to avoid pushing the same neuron twice
	seen bool
}

func newSensorNeuron() *neuron {
	return _newNeuron(sensorNeuron)
}

func newOutputNeuron() *neuron {
	return _newNeuron(outputNeuron)
}

func newHiddenNeuron() *neuron {
	return _newNeuron(hiddenNeuron)
}

func _newNeuron(kind neuronKind) *neuron {
	return &neuron{
		id: neuronID(nextID()),
		innovation: nextInnovation(),
		kind: kind,
	}
}

func (n *neuron) clone() *neuron {
	copyNeuron := *n
	return &copyNeuron
}

func (n *neuron) getInnovation() uint64 {
	return n.innovation
}

// An organism that holds a set of neurons and synapses.
type organism struct {
	// A set of sensor neurons
	sensors []neuronID
	// A set of output neurons
	outputs []neuronID
	// Map from neuron id to neuron
	neurons map[neuronID]*neuron
	// Map from synapse id to synapse
	synapses map[synapseID]*synapse
	// Map from neuron id to outgoing synapse ids
	connections map[neuronID][]synapseID

	// Genes in order of appearance
	genes []gene

	// Generation = parent generation + 1
	generation int

	// Evolutionary fitness value
	fitness float64
}

type species struct {
	population []organism
}

// Creates an empty organism
func _newOrganism(nInputs, nOutputs int) *organism {
	sensors := make([]neuronID, nInputs)
	sensors = sensors[:0]
	outputs := make([]neuronID, nOutputs)
	outputs = outputs[:0]

	neurons := make(map[neuronID]*neuron)
	synapses := make(map[synapseID]*synapse)
	connections := make(map[neuronID][]synapseID)
	genes := make([]gene, nInputs * nOutputs)
	genes = genes[:0]

	return &organism{
		sensors:  sensors,
		outputs:  outputs,
		neurons:  neurons,
		synapses: synapses,
		connections: connections,
		genes: genes,
	}
}

func newOrganism(nInputs, nOutputs int) *organism {
	org := _newOrganism(nInputs, nOutputs)

	// Create sensor neurons
	for i := 0; i < nInputs; i++ {
		org.addNeuron(newSensorNeuron())
	}

	// Create output neurons
	for i := 0; i < nOutputs; i++ {
		org.addNeuron(newOutputNeuron())
	}

	// Match sensors with outputs
	for i := 0; i < max(nInputs, nOutputs); i++ {
		in := org.neurons[org.sensors[i % nInputs]]
		out := org.neurons[org.outputs[i % nOutputs]]

		org.addSynapse(newSynapse(in, out))
	}

	return org
}

func (org *organism) clone() *organism {
	clone := _newOrganism(len(org.sensors), len(org.outputs))

	for _, gene := range org.genes {
		switch g := gene.(type) {
		case *neuron:
			clone.addNeuron(g.clone())
		case *synapse:
			clone.addSynapse(g.clone())
		}
	}

	return clone
}

// Add a neuron
func (org *organism) addNeuron(neuron *neuron) {
	org.neurons[neuron.id] = neuron
	org.genes = append(org.genes, neuron)

	switch neuron.kind {
	case sensorNeuron:
		org.sensors = append(org.sensors, neuron.id)
	case outputNeuron:
		org.outputs = append(org.outputs, neuron.id)
	}
}

// Add a new synapse
func (org *organism) addSynapse(synapse *synapse) {
	org.synapses[synapse.id] = synapse
	org.connections[synapse.in] = append(org.connections[synapse.in], synapse.id)
	org.genes = append(org.genes, synapse)
}

// Lookup a neuron
func (org *organism) getNeuron(id neuronID) *neuron {
	return org.neurons[id]
}

// Lookup a synapse
func (org *organism) getSynapse(id synapseID) *synapse {
	return org.synapses[id]
}

// Lookup synapse endpoint neurons
func (org *organism) synapseEndpoints(id synapseID) (*neuron, *neuron) {
	synapse := org.getSynapse(id)
	return org.neurons[synapse.in], org.neurons[synapse.out]
}

// Mutate the organism
func (org *organism) mutate() {
	for _, synapseIDs := range org.connections {
		for _, id := range synapseIDs {
			// Instead of just doing everything there we delegate, this
			// makes testing a lot easier

			if RandFloat64() <= config.OrganismConfig.SynapseSplitMutProb {
				org.splitSynapse(id)
			}
			if RandFloat64() <= config.OrganismConfig.SynapseActivityMutProb {
				org.toggleEnabled(id)	
			}

			if RandFloat64() <= config.OrganismConfig.SynapseWeightMutProp {
				org.mutateWeight(id)
			}
		}
	}
}

// Split a synapse, creates two new synapses with a neuron in between
// to replace the old synapse and then disables the old synapse.
func (org *organism) splitSynapse(id synapseID) {

	// The in and out neurons of this synapse
	in, out := org.synapseEndpoints(id)

	// The new neuron
	neuron := newHiddenNeuron()

	// A new synapse from the in neuron to the new neuron
	synIn := newSynapse(in, neuron)
	// A new synapse from the new neuron to the out neuron
	synOut := newSynapse(neuron, out)

	// The replaced synapse becomes inactive
	org.synapses[id].enabled = false

	// Now do the bookkeeping in the organism, it is important that
	// the genes are added in order and that the neuron is added
	// before any synapses referencing it
	org.addNeuron(neuron)
	org.addSynapse(synIn)
	org.addSynapse(synOut)
}

func (org *organism) toggleEnabled(id synapseID) {
	org.synapses[id].toggleEnabled()
}

func (org *organism) mutateWeight(id synapseID) {
	org.synapses[id].mutateWeight()
}

// The "genetic distance" between two organism
type distance struct {
	// The number of excess genes
	excess int
	// The number of disjoint genes
	disjoint int
	// The average weight difference of matching genes
	weightDiff float64
	// The number of genes in the larger genome
	nbrGenes int
}

// Mate two organism producing an offspring with the combined topology
// of its parents.
func mate(a, b *organism) *organism {
	if len(a.sensors) != len(b.sensors) ||
		len(a.outputs) != len(b.outputs) {
		log.Fatal("Wooooh easy there fella' that's illegal")
	}

	// Create an empty offspring
	offspring := _newOrganism(len(a.sensors), len(a.outputs))
	offspring.generation = a.generation + 1

	// Line up the genes and start building the new topology
	aLen := len(a.genes) // Number of genes in a
	bLen := len(b.genes) // and in b
	
	var aIdx int // the index into a.innovation
	var bIdx int // the index into b.innovation

	// Start copying the genes into the offspring
	for aIdx, bIdx  = 0, 0; aIdx < aLen || bIdx < bLen; {
		var aGene gene
		var bGene gene

		// Get the next gene from the parents unless they're exhausted
		if aIdx < aLen {
			aGene = a.genes[aIdx]
		}
		if bIdx < bLen {
			bGene = b.genes[bIdx]
		}

		// This is what the child will inherit
		var inheritance gene

		// Both parent could provide genes
		if aGene != nil && bGene != nil {

			aInov := aGene.getInnovation()
			bInov := bGene.getInnovation()

			if aInov == bInov {
				// If these are the same genes inherit from the fittest parent
				if a.fitness > b.fitness {
					inheritance = aGene
				} else {
					inheritance = bGene
				}

				aIdx++
				bIdx++
			} else if aInov < bInov {
				// Inherit from a if it has the lower innovation number
				inheritance = aGene
				aIdx++
			} else if bInov < aInov {
				// Inherit from b if it has the lower innovation number
				inheritance = bGene
				bIdx++
			}

		} else if aGene != nil {

			// The end of b's genes has been reached, inherit from a
			inheritance = aGene
			aIdx++

		} else if bGene != nil {

			// The end of a's genes has been reached, inherit from b
			inheritance = bGene
			bIdx++

		} else {
			log.Fatal("Out of genes but haven't reached end of genes")
		}

		// Now insert the inherited gene into the offspring
		switch g := inheritance.(type) {

		case *neuron:
			copyNeuron := *g
			offspring.addNeuron(&copyNeuron)
		case *synapse:
			copySynapse := *g
			offspring.addSynapse(&copySynapse)
		}
	}

	return offspring
}

// Feed a new slice of inputs to the organism
func (org *organism) process(input []float64) []float64 {
	if len(input) != len(org.sensors) {
		log.Fatal("Number of inputs exceeds number of sensors")
	}

	// Clear all neurons
	for _, neuron := range org.neurons {
		// Set the current sum equal to the recursive inputs from
		// the previous iteration
		neuron.sum, neuron.future = neuron.future, 0

		neuron.visited = false
		neuron.seen = false
	}

	// Add the input signals to the sensor neurons
	for i, id := range org.sensors {
		s := org.neurons[id]
		s.sum += input[i]
	}

	org.propagate()

	out := make([]float64, len(org.outputs))
	for i, id := range org.outputs {
		out[i] = org.neurons[id].value
	}

	return out
}

// Propagate signals through the organismt network toplogy
func (org *organism) propagate() {
	// Queue used for breadth first traversal of the network
	queue := newsqueue()

	// Start by adding the input neurons to the queue
	for _, id := range org.sensors {
		queue.Push(org.neurons[id])
	}

	// Iterate as long as there are unprocessed nueurons in the queue
	for queue.Size() > 0 {

		// Pop the queue
		n := queue.Pop().(*neuron)

		// This neuron has already been traversed...
		if n.visited {
			// ...and this situation cannot occur unless there's a bug
			log.Fatal("Found visited neuron, ", n, ", in the queue")
		}

		// Tag the neuron as visited and calculate the output value
		n.visited = true
		n.value = config.OrganismConfig.actFunc(n.sum)

		// Propagate the output value through the synapses
		for _, id := range org.connections[n.id] {
			synapse := org.getSynapse(id)

			if synapse.enabled {

				signal := n.value * synapse.weight
				out := org.neurons[synapse.out]

				if out.visited {
					// If the attached neuron has already been visited then
					// this is a recurrent network and we store the value
					// to be processed at the next input iteration
					out.future += signal
				} else {
					// The attached neuron hasn't been traveresed yet. Add
					// the signal to the input sum and push the neuron onto
					// the queue.
					out.sum += signal

					// This is part of the breadth first traversal, avoid pushing
					// the same neuron twice.
					if !out.seen {
						out.seen = true
						queue.Push(out)
					}
				}
			}
		}
	}
}
