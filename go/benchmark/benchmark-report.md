# Run time of NetLogo, gobees and gobees-fast

The example farm with 6 seeds, 365 days each, 2 runs at the same time in all models. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI switches off. Details: go/benchmark/README.md.

## NetLogo and gobees-fast

Setup is the time of NetLogo's Setup and of the gobees-fast setup, including reading the input files. The days are timed separately, without output during the run.

| | NetLogo median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast |
| --- | ---: | ---: | ---: |
| Setup | 636 ms (345 ms-3.2 s) | 5 ms (3 ms-6 ms) | 124 |
| 365 days | 30.1 s (24.0 s-41.8 s) | 318 ms (276 ms-371 ms) | 95 |

## NetLogo, gobees and gobees-fast

gobees is the base commit 40dc301, gobees-fast the current code. Both run as command line programs without output, one version after the other, and are timed as a whole, from the start of the program to its end, including reading the input files. NetLogo's whole run is its Setup and its days, without the start of Java.

| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast | gobees / gobees-fast |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole run | 30.7 s (24.4 s-45.0 s) | 325 ms (269 ms-383 ms) | 332 ms (264 ms-382 ms) | 93 | 0.98 |

## Runs

The runs of NetLogo and of the Go versions with the same seed are not the same run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a comparable amount of work.

| Run | Seed | NetLogo setup | NetLogo days | NetLogo colonies founded | gobees-fast setup | gobees-fast days | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 3.2 s | 41.8 s | 154 | 6 ms | 297 ms | 132 |
| 2 | 2 | 3.2 s | 32.7 s | 150 | 6 ms | 276 ms | 123 |
| 3 | 3 | 690 ms | 30.2 s | 124 | 4 ms | 355 ms | 140 |
| 4 | 4 | 583 ms | 29.8 s | 127 | 6 ms | 312 ms | 138 |
| 5 | 5 | 563 ms | 30.1 s | 144 | 3 ms | 324 ms | 120 |
| 6 | 6 | 345 ms | 24.0 s | 130 | 4 ms | 371 ms | 141 |

As programs:

| Run | Seed | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 310 ms | 132 | 303 ms | 132 |
| 2 | 2 | 269 ms | 123 | 264 ms | 123 |
| 3 | 3 | 356 ms | 140 | 366 ms | 140 |
| 4 | 4 | 314 ms | 138 | 332 ms | 138 |
| 5 | 5 | 337 ms | 120 | 332 ms | 120 |
| 6 | 6 | 383 ms | 141 | 382 ms | 141 |

## Machine

- CPU: INTEL(R) XEON(R) PLATINUM 8573C
- Logical cores: 2
- OS: linux/amd64
- Go: go1.26.5
- Java: openjdk version "21.0.13" 2026-10-20 LTS
- Commit: 06e2ad6
