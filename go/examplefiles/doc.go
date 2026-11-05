// Package examplefiles holds the NetLogo example input files, or their beginning, copied verbatim.
//
// Import tests parse these strings and check every value, so they do not depend on the files in netlogo/.
// Do not edit the strings by hand: copy them again from the files.
//
// The CSV files have Windows line endings (CRLF). Go raw strings drop the carriage return, so the copies
// use LF line endings.
// The tests also parse each copy with CRLF line endings.
package examplefiles
