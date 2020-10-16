// +build !race

package httptestbench

// RaceDetectorEnabled exposes Go race detector status.
//
// Race detector affects performance and allocations so it has to be taken in account when asserting benchmark results.
const RaceDetectorEnabled = false
