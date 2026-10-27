# Run time of gobees and NetLogo

The example farm with 6 seeds, 365 days each, 2 runs at the same time in both models. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off. Setup is the time of NetLogo's Setup and of the gobees setup, including reading the input files. The days are timed separately, without output during the run. Details: go/benchmark/README.md.

| | NetLogo median (min-max) | gobees median (min-max) | NetLogo / gobees |
| --- | ---: | ---: | ---: |
| Setup | 646 ms (379 ms-2.6 s) | 5 ms (3 ms-10 ms) | 121 |
| 365 days | 31.3 s (25.3 s-43.2 s) | 359 ms (282 ms-403 ms) | 87 |

## Runs

The runs of both models with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that both did a comparable amount of work.

| Run | Seed | NetLogo setup | NetLogo days | NetLogo colonies founded | gobees setup | gobees days | gobees colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 2.6 s | 43.2 s | 154 | 10 ms | 348 ms | 132 |
| 2 | 2 | 2.6 s | 32.2 s | 150 | 7 ms | 282 ms | 123 |
| 3 | 3 | 680 ms | 31.4 s | 124 | 4 ms | 401 ms | 140 |
| 4 | 4 | 612 ms | 29.6 s | 127 | 3 ms | 366 ms | 138 |
| 5 | 5 | 379 ms | 31.3 s | 144 | 6 ms | 351 ms | 120 |
| 6 | 6 | 503 ms | 25.3 s | 130 | 5 ms | 403 ms | 141 |

## Machine

- CPU: AMD EPYC 7763 64-Core Processor
- Logical cores: 2
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.12.1" 2026-08-18 LTS
- Commit: 070c688
