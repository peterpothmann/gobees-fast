# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. gobees is the base commit ff06b97, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run as command line programs without output, one version after the other, and are timed from the start of the program to its end, including reading the input files. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 14.0 s (12.4 s-18.4 s) | 185 ms (144 ms-226 ms) | 61 ms (49 ms-64 ms) | 228 | 3.02 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 18.4 s | 154 | 165 ms | 132 | 64 ms | 142 |
| 2 | 2 | 14.7 s | 150 | 144 ms | 123 | 49 ms | 157 |
| 3 | 3 | 13.7 s | 124 | 203 ms | 140 | 60 ms | 149 |
| 4 | 4 | 13.2 s | 127 | 180 ms | 138 | 64 ms | 153 |
| 5 | 5 | 14.3 s | 144 | 190 ms | 120 | 56 ms | 145 |
| 6 | 6 | 12.4 s | 130 | 226 ms | 141 | 63 ms | 125 |

## Machine

- CPU: Intel(R) Xeon(R) 6975P-C
- Logical cores: 4
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: a293765
