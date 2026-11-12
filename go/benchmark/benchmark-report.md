# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 16.7 s (15.2 s-24.1 s) | 267 ms (207 ms-322 ms) | 82 ms (71 ms-92 ms) | 203 | 3.26 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 24.1 s | 154 | 239 ms | 132 | 92 ms | 142 |
| 2 | 2 | 19.4 s | 150 | 207 ms | 123 | 71 ms | 157 |
| 3 | 3 | 17.0 s | 124 | 286 ms | 140 | 81 ms | 149 |
| 4 | 4 | 16.2 s | 127 | 263 ms | 138 | 92 ms | 153 |
| 5 | 5 | 16.3 s | 144 | 272 ms | 120 | 77 ms | 145 |
| 6 | 6 | 15.2 s | 130 | 322 ms | 141 | 83 ms | 125 |

## Machine

- CPU: AMD EPYC 7763 64-Core Processor
- Logical cores: 4
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: 9ca7cf1
