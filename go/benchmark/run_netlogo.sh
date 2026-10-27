#!/usr/bin/env bash
# Runs the benchmark experiment go/benchmark/netlogo.xml headless with NetLogo 5.3.1.
#
# Usage (from the repository root):
#   go/benchmark/run_netlogo.sh <NetLogo folder> <threads> <table.csv> <console.txt>
#
# <NetLogo folder> holds app/NetLogo.jar, e.g. netlogo-5.3.1-64 from NetLogo-5.3.1-64.tgz (Linux) or
# "/c/Program Files/NetLogo 5.3.1" (Windows, Git Bash). Java is taken from the PATH, or from $JAVA.
# <threads> runs are made at the same time. The table gets the metrics at the end of each run, the console output
# (also shown) holds the setup times.
# NetLogo finds its extensions (profiler, csv) only if the working directory is its app folder, so the script
# changes into it. The model file is not changed: the experiment comes from a separate setup file.
set -euo pipefail

if [ $# -ne 4 ]; then
    echo "usage: $0 <NetLogo folder> <threads> <table.csv> <console.txt>" >&2
    exit 2
fi
netlogo=$(realpath "$1")
threads=$2
table=$(realpath -m "$3")
console=$(realpath -m "$4")
model=$(realpath netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo)
experiment=$(realpath go/benchmark/netlogo.xml)

# classpath as in NetLogo.cfg (do NOT add all jars: NetLogoLite.jar contains conflicting classes)
separator=":"
case "$(uname -s)" in MINGW* | MSYS* | CYGWIN*) separator=";" ;; esac
cd "$netlogo/app"
classpath="NetLogo.jar${separator}$(sed -n 's/^app\.classpath=//p' NetLogo.cfg | tr -d '\r' | tr ':;' "${separator}${separator}")"

"${JAVA:-java}" -version 2>&1 | head -n 1
start=$(date +%s)
# java.awt.headless: the CI runner has no display
"${JAVA:-java}" -Xmx4096m -Dfile.encoding=UTF-8 -Djava.awt.headless=true -cp "$classpath" org.nlogo.headless.Main \
    --model "$model" --setup-file "$experiment" --experiment benchmark --table "$table" --threads "$threads" |
    tee "$console"
echo "Done after $(($(date +%s) - start)) s"
