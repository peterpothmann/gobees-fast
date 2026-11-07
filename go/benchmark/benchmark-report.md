# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 23.2 s (19.6 s-33.8 s) | 257 ms (202 ms-306 ms) | 198 ms (158 ms-230 ms) | 117 | 1.30 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 33.8 s | 154 | 257 ms | 132 | 176 ms | 132 |
| 2 | 2 | 27.9 s | 150 | 202 ms | 123 | 158 ms | 123 |
| 3 | 3 | 22.9 s | 124 | 294 ms | 140 | 203 ms | 140 |
| 4 | 4 | 22.1 s | 127 | 257 ms | 138 | 193 ms | 138 |
| 5 | 5 | 23.5 s | 144 | 249 ms | 120 | 203 ms | 120 |
| 6 | 6 | 19.6 s | 130 | 306 ms | 141 | 230 ms | 141 |

## Machine

- CPU: Intel(R) Xeon(R) 6973P-C
- Logical cores: 2
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: ce5c871
