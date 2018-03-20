package neat

import (
	"os"
	"testing"
	"github.com/stretchr/testify/require"
)

var identity = func(a float64) float64 { return a }

var testConfig = NeatConfig{
	SpeciesConfig: SpeciesConfig{
		ExcessGenesCoeff: 0.1,
		DisjoinGenesCoeff: 0.2,
		AvgWeightDiffCoeff: 0.1,
		CompatibilityThreshold: 0.5,
	},
	OrganismConfig: OrganismConfig{
		SynapseSplitMutProb: 0.01,
		SynapseActivityMutProb: 0.01,
		SynapseWeightMutProp: 0.01,
		SynapseWeightBound: 5.0,
		actFunc: identity,
	},
}

// Generate a simple recurrent network
// +--------+   +--------+   +--------+
// | Sensor |---| Hidden |---| Output |
// +--------+   +--------+   +--------+
//     |            |
//     +------------+
func createSimpleRecurrent() *organism {
	org := _newOrganism(1, 1)

	sensor := newSensorNeuron()
	hidden := newHiddenNeuron()
	output := newOutputNeuron()

	org.addNeuron(sensor)
	org.addNeuron(hidden)
	org.addNeuron(output)

	org.addSynapse(newSynapse(sensor, hidden))
	org.addSynapse(newSynapse(hidden, sensor))
	org.addSynapse(newSynapse(hidden, output))

	return org
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	SetNeatConfig(testConfig)
	result := m.Run()
	os.Exit(result)
}

// Test a simple network with one input and one output. The input
// should produce identical output.
func TestSimple(t *testing.T) {
	nInputs := 1
	nOutputs := 1
	organism := newOrganism(nInputs, nOutputs)

	value := float64(1.0)
	input := []float64{value}
	output := organism.process(input)

	require.Equal(t, output[0], value, "")
}

// Test a network with one input and two outputs, The input
// should produce identical output at both outputs.
func TestSimple2(t *testing.T) {
	nInputs := 1
	nOutputs := 2
	organism := newOrganism(nInputs, nOutputs)

	value := float64(1.0)
	input := []float64{value}
	output := organism.process(input)

	require.Equal(t, output[0], value, "")
	require.Equal(t, output[1], value, "")
}

// Test a network with two inputs and one output, The output
// should be equals to the sum of the inputs.
func TestSimple3(t *testing.T) {
	nInputs := 2
	nOutputs := 1
	organism := newOrganism(nInputs, nOutputs)

	value := float64(1.0)
	input := []float64{value, value}
	output := organism.process(input)

	require.Equal(t, output[0], 2 * value, "")
}

// Test a simple recurrent network where the sensor
// output is fed back to the sensor. The output is
// the sum of the input values over time.
//
// Input: 1 0 0 0 0, gives
// Output: 1 1 1 1 1
func TestSimpleRecurrent(t *testing.T) {
	org := createSimpleRecurrent()

	// Test this sequence
	// In: 1 0 0 1, gives
	// Out: 1 1 1 2
	ioAll := [][][]float64{
		[][]float64{[]float64{1}, []float64{1}},
		[][]float64{[]float64{0}, []float64{1}},
		[][]float64{[]float64{0}, []float64{1}},
		[][]float64{[]float64{1}, []float64{2}},
	}

	for _, io := range ioAll {
		input := io[0]
		refOut := io[1]

		out := org.process(input)
		require.Equal(t, out, refOut, "")
	}
}

func TestSplitSynapse(t *testing.T) {
	// Set up a minimal network

	//
	// +--------+              +--------+
	// | Sensor |---Synapse1---| Output |
	// +--------+              +--------+
	//

	organism := newOrganism(1, 1)

	// Dig out the synapses connecting the sensor to the output
	synapse1 := func() *synapse {
		for _, s := range organism.synapses {
			return s
		}
		panic("Can't happen")
	}()

	sensor, output := organism.synapseEndpoints(synapse1.id)
	organism.splitSynapse(synapse1.id)

	// Assert that the split synapse has been disabled
	require.Equal(t, synapse1.enabled, false, "")

	// Now we expect the network to look like this
	//
	// +--------+              +--------+              +--------+
	// | Sensor |---Synapse2---| Hidden |---Synapse3---| Output |
	// +--------+              +--------+              +--------+
	//     |                                               |
	//     +----------------Synapse1(disabled)-------------+
	//

	// The second synapse among the sensor synapses is synapse2
	synapse2 := organism.getSynapse(organism.connections[sensor.id][1])
	hidden := organism.getNeuron(synapse2.out)
	synapse3 := organism.getSynapse(organism.connections[hidden.id][0])

	require.Equal(t, synapse3.out, output.id, "")
}

func TestOrganismClone(t *testing.T) {
	a := createSimpleRecurrent()
	b := a.clone()

	// Test this sequence
	// In: 1 0 0 1, gives
	// Out: 1 1 1 2
	ioAll := [][][]float64{
		[][]float64{[]float64{1}, []float64{1}},
		[][]float64{[]float64{0}, []float64{1}},
		[][]float64{[]float64{0}, []float64{1}},
		[][]float64{[]float64{1}, []float64{2}},
	}

	for _, io := range ioAll {
		input := io[0]
		refOut := io[1]

		out := a.process(input)
		require.Equal(t, out, refOut, "")

		out = b.process(input)
		t.Log("in: ", input, " refOut: ", refOut, " out: ", out)
		require.Equal(t, out, refOut, "")
	}
}

func TestMating(t *testing.T) {
	a := newOrganism(2, 2)
	b := a.clone()

	offspring := mate(a, b)

	t.Log(offspring)
}
