# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 23.6 s (19.2 s-35.1 s) | 299 ms (229 ms-317 ms) | 143 ms (137 ms-148 ms) | 164 | 2.08 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 35.1 s | 154 | 251 ms | 132 | 148 ms | 144 |
| 2 | 2 | 29.7 s | 150 | 229 ms | 123 | 137 ms | 143 |
| 3 | 3 | 23.6 s | 124 | 309 ms | 140 | 146 ms | 128 |
| 4 | 4 | 22.9 s | 127 | 305 ms | 138 | 146 ms | 150 |
| 5 | 5 | 23.6 s | 144 | 292 ms | 120 | 141 ms | 145 |
| 6 | 6 | 19.2 s | 130 | 317 ms | 141 | 141 ms | 144 |

## Machine

- CPU: AMD EPYC 9V74 80-Core Processor
- Logical cores: 2
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: a9a2370
