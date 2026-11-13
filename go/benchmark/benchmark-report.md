# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 9.4 s (8.2 s-13.8 s) | 160 ms (122 ms-178 ms) | 51 ms (43 ms-60 ms) | 185 | 3.16 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 13.8 s | 154 | 142 ms | 132 | 60 ms | 142 |
| 2 | 2 | 11.0 s | 150 | 122 ms | 123 | 48 ms | 157 |
| 3 | 3 | 9.5 s | 124 | 168 ms | 140 | 53 ms | 149 |
| 4 | 4 | 9.0 s | 127 | 174 ms | 138 | 57 ms | 153 |
| 5 | 5 | 9.2 s | 144 | 153 ms | 120 | 43 ms | 145 |
| 6 | 6 | 8.2 s | 130 | 178 ms | 141 | 48 ms | 125 |

## Machine

- CPU: AMD EPYC 9V45 96-Core Processor
- Logical cores: 4
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: bd36658
