
This tool runs a set of matching simulation tests, extracts stats from their output and generates a csv to compare them easily.

Note: The parsing logic might break in the future if the `run_matching_simulator.sh` starts spitting different shaped lines. Alternative is to load all the event logs into a sqlite table and then run queries on top instead of parsing outputs of jq in this tool.


Run all the scenarios and compare:
```
go run tools/matchingsimulationcomparison/*.go --out simulation_comparison.csv
```

If you have already run some scenarios before and made changes in the output/comparison then run in Compare mode
```
go run tools/matchingsimulationcomparison/*.go --out simulation_comparison.csv \
    --ts 2024-09-12-18-16-44 \
    --mode Compare
```
