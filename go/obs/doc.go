// Package obs holds observers, which read model output from the world. They implement the observer interfaces of
// ark-tools and are written to files by the reporters of ark-tools, e.g. reporter.CSV.
//
// The model holds only state that its dynamics read. The observers compute the output from this state,
// and count from Ark events what the state does not hold (Counters).
package obs
