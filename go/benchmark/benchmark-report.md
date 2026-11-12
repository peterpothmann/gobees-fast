# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 17.6 s (15.2 s-27.8 s) | 265 ms (209 ms-317 ms) | 131 ms (119 ms-141 ms) | 134 | 2.03 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 27.8 s | 154 | 245 ms | 132 | 141 ms | 144 |
| 2 | 2 | 22.4 s | 150 | 209 ms | 123 | 119 ms | 143 |
| 3 | 3 | 17.7 s | 124 | 312 ms | 140 | 130 ms | 128 |
| 4 | 4 | 17.1 s | 127 | 274 ms | 138 | 141 ms | 150 |
| 5 | 5 | 17.4 s | 144 | 257 ms | 120 | 132 ms | 145 |
| 6 | 6 | 15.2 s | 130 | 317 ms | 141 | 129 ms | 144 |

## Machine

- CPU: AMD EPYC 7763 64-Core Processor
- Logical cores: 4
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: c2f5626
