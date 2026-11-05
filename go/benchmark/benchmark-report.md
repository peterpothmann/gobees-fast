# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 32.8 s (27.3 s-50.0 s) | 368 ms (290 ms-471 ms) | 292 ms (258 ms-325 ms) | 112 | 1.26 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 50.0 s | 154 | 319 ms | 132 | 276 ms | 132 |
| 2 | 2 | 41.3 s | 150 | 290 ms | 123 | 258 ms | 123 |
| 3 | 3 | 31.8 s | 124 | 384 ms | 140 | 312 ms | 140 |
| 4 | 4 | 32.1 s | 127 | 351 ms | 138 | 285 ms | 138 |
| 5 | 5 | 33.4 s | 144 | 403 ms | 120 | 298 ms | 120 |
| 6 | 6 | 27.3 s | 130 | 471 ms | 141 | 325 ms | 141 |

## Machine

- CPU: AMD EPYC 7763 64-Core Processor
- Logical cores: 2
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: d95844c
