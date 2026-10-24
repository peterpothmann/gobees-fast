# Runs a BehaviorSpace experiment of BEE-STEWARD headless with NetLogo 5.3.1.
#
# Usage (PowerShell, from the repository root):
#   .\go\validation\scripts\run_netlogo_experiment.ps1 -Experiment go\validation\experiments\step3.xml -Output step3_netlogo.csv
#
# NetLogo 5.3.1 ships without java.exe, so the Java on the PATH is used (tested with Java 21).
# NetLogo finds its extensions (profiler, csv) only if the working directory is its app folder, so the script
# changes into it. The model file is not changed: the experiment comes from a separate setup file.

param(
    [Parameter(Mandatory = $true)][string]$Experiment,
    [Parameter(Mandatory = $true)][string]$Output,
    [int]$Threads = [Environment]::ProcessorCount,
    [string]$NetLogoDir = "C:\Program Files\NetLogo 5.3.1",
    [string]$Model = "netlogo\__BEE-STEWARD_RC_ColourVersionMB.nlogo",
    [string]$Memory = "4096m"
)

$ErrorActionPreference = "Stop"
$experimentPath = (Resolve-Path $Experiment).Path
$modelPath = (Resolve-Path $Model).Path
if ([System.IO.Path]::IsPathRooted($Output)) { $outputPath = $Output }
else { $outputPath = [System.IO.Path]::GetFullPath((Join-Path (Get-Location) $Output)) }

# the experiment name is the name attribute of the first <experiment> in the setup file
$name = ([xml](Get-Content $experimentPath)).experiments.experiment[0].name
if (-not $name) { $name = ([xml](Get-Content $experimentPath)).experiments.experiment.name }

# classpath as in NetLogo.cfg (do NOT add all jars: NetLogoLite.jar contains conflicting classes)
$app = Join-Path $NetLogoDir "app"
$cfg = Get-Content (Join-Path $app "NetLogo.cfg") | Where-Object { $_ -like "app.classpath=*" }
$classpath = "NetLogo.jar;" + $cfg.Substring("app.classpath=".Length)

Write-Host "Experiment $name, $Threads threads, output $outputPath"
$start = Get-Date
Push-Location $app
try {
    & java "-Xmx$Memory" "-Dfile.encoding=UTF-8" -cp $classpath org.nlogo.headless.Main `
        --model $modelPath --setup-file $experimentPath --experiment $name --table $outputPath --threads $Threads
    if ($LASTEXITCODE -ne 0) { throw "NetLogo exited with code $LASTEXITCODE" }
}
finally {
    Pop-Location
}
Write-Host ("Done after {0:N0} min" -f ((Get-Date) - $start).TotalMinutes)
