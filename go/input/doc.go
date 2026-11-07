// Package input reads a run file (YAML) and its landscape files into params.Params before the setup [INIT-04],
// [INIT-05]. No system opens a file afterwards.
//
// The tool tools/netlogo2gobees converts the NetLogo input files into run files. gobees-fast reads no NetLogo files
// since step F7.
package input
