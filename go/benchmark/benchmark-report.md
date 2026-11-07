# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 24.7 s (20.3 s-35.0 s) | 287 ms (224 ms-368 ms) | 230 ms (201 ms-258 ms) | 107 | 1.25 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 35.0 s | 154 | 280 ms | 132 | 215 ms | 132 |
| 2 | 2 | 29.5 s | 150 | 224 ms | 123 | 201 ms | 123 |
| 3 | 3 | 24.9 s | 124 | 368 ms | 140 | 250 ms | 140 |
| 4 | 4 | 23.6 s | 127 | 287 ms | 138 | 227 ms | 138 |
| 5 | 5 | 24.4 s | 144 | 287 ms | 120 | 233 ms | 120 |
| 6 | 6 | 20.3 s | 130 | 315 ms | 141 | 258 ms | 141 |

## Machine

- CPU: AMD EPYC 9V74 80-Core Processor
- Logical cores: 2
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: ce10fd5
