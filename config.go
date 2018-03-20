package neat

import (
	"errors"
	"encoding/json"
	"math"
	"io/ioutil"
)

var ErrIllegalProbability = errors.New("probability is not in range [0, 1]")
var ErrNoSuchFunction = errors.New("No such function")

// The sigmoid function conveniently available
func Sigmoid(x float64) float64 {
	expX := math.Exp(x)
	return expX / (expX + 1)
}

// The fast "sigmoid"
func FastSigmoid(x float64) float64 {
	return x / (1 + math.Abs(x))
}

// The rectifier function, see
// https://en.wikipedia.org/wiki/Rectifier_(neural_networks)
func Rectifier(x float64) float64 {
	return math.Max(0, x)
}

var actFuncNameMap = map[string]ActivationFunction{
	"Sigmoid": Sigmoid,
	"FastSigmoid": FastSigmoid,
	"Recifier": Rectifier,
}

type SpeciesConfig struct {
	/*
	// When a organism evolves a new topology it may need to be treated as a
	// new species in order to let its new topology mature a bit before being
	// subjected to full competition. Set the maximum number of generations
	// that the new species may evolve under before being deemed unfit.
	MaxAdaptationGenerations int `json:"MaxAdaptationGenerations"`

	// A population cannot be allowed to grow inifitely large. Cap the
	// number of organism in population after mating.
	MaxOrganismInPopulation int `json:"MaxOrganismInPopulation"`
	*/


	// The three following coefficients are used to calculate the genetic distance
	// between two organisms.
	//
	// d = (c1 * E  + c2 * D) / N + c3 * W
	//
	// Where
	// - c1 is the ExcessGeneCoeff
	// - E is the number of excess genes
	// - c2 is the DisjoinGenesCoeff
	// - D is the number of disjoint genes
	// - c3 is the AvgWeightDiffCoeff
	// - W is the average weight difference of matching genes
	// - N is the number of genes in the larger genome

	// Excess genes coefficient
	ExcessGenesCoeff float64 `json:"ExcessGenesCoeff"`

	// Disjoint genes coefficient
	DisjoinGenesCoeff float64 `json:"DisjoinGenesCoeff"`

	// Average weight diff coefficient
	AvgWeightDiffCoeff float64 `json:"AvgWeightDiffCoeff "`

	// The compatibility threshold, i.e. the maximum genetic distance
	// separating two organisms before speciation occurs.
	CompatibilityThreshold float64 `json:"CompatibilityThreshold"`
}

type OrganismConfig struct {
	// The probablility that a synapse is split
	SynapseSplitMutProb float64 `json:"SynapseSplitMutProb"`

	// The probability that a synapse's activity toggle is switched
	SynapseActivityMutProb float64 `json:"SynapseActivityMutProb"`

	// The probability that a synapse's weight is perturbed
	SynapseWeightMutProp float64 `json:"SynapseWeightMutProp"`

	// The absolute bound of a weight mutation (rand-number * bound)
	SynapseWeightBound float64 `json:"SynapseWeightBound"`

	// Neuron activation function
	ActFunc string `json:"ActFunc"`

	actFunc ActivationFunction
}

type NeatConfig struct {
	SpeciesConfig SpeciesConfig `json:"SpeciesConfig"`
	OrganismConfig OrganismConfig `json:"OrganismConfig"`
}

func validateSpeciesConfig(c SpeciesConfig) error {
	if c.ExcessGenesCoeff < 0 {
		errors.New("ExcessGeneCoeff must be positive")
	}

	if c.DisjoinGenesCoeff < 0 {
		errors.New("DisjoinGenesCoeff must be positive")
	}

	if c.AvgWeightDiffCoeff <0 {
		errors.New("AvgWeightDiffCoeff must be positive")
	}

	if c.CompatibilityThreshold < 0 {
		errors.New("CompatibilityThreshold must be positive")
	}

	return nil
}

func validateOrganismConfig(c OrganismConfig) error {
	if !inRange(c.SynapseSplitMutProb, 0.0, 1.0) {
		return errors.New("SynapseSplitMutProb must be in the range [0, 1]")
	}

	if !inRange(c.SynapseActivityMutProb, 0.0, 1.0) {
		return errors.New("SynapseActivityMutProb must be in the range [0, 1]")
	}

	if !inRange(c.SynapseWeightMutProp, 0.0, 1.0) {
		return errors.New("SynapseWeightMutProp must be in the range [0, 1]")
	}

	if c.SynapseWeightBound <= 0 {
		return errors.New("SynapseWeightBound must be larger than zero")
	}

	if _, ok := actFuncNameMap[c.ActFunc]; !ok {
		return errors.New("Unregistered activation function: " + c.ActFunc)
	}

	return nil
}

func validateNeatConfig(c NeatConfig) error {
	if err := validateSpeciesConfig(c.SpeciesConfig); err != nil {
		return err
	}

	if err := validateOrganismConfig(c.OrganismConfig); err != nil {
		return err
	}

	return nil
}

func ReadConfig(path string) (*NeatConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config NeatConfig
	err = json.Unmarshal(raw, &config)

	if err = validateNeatConfig(config); err != nil {
		return nil, err
	}
	
	return &config, err
}
