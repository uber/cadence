#!/bin/bash

# This script can be used to run matching simulator and check the critical flow via logs
#

set -eo pipefail

testCase="${1:-default}"
testCfg="testdata/matching_simulation_$testCase.yaml"
now="$(date '+%Y-%m-%d-%H-%M-%S')"
timestamp="${2:-$now}"
testName="test-$testCase-$timestamp"
resultFolder="matching-simulator-output"
mkdir -p "$resultFolder"
eventLogsFile="$resultFolder/events.json"
testSummaryFile="$resultFolder/$testName-summary.txt"

echo "Building test image"
docker-compose -f docker/buildkite/docker-compose-local-matching-simulation.yml \
  build matching-simulator

echo "Running the test $testCase"
docker-compose \
  -f docker/buildkite/docker-compose-local-matching-simulation.yml \
  run -e MATCHING_SIMULATION_CONFIG=$testCfg --rm --remove-orphans matching-simulator \
  | grep -a --line-buffered "Matching New Event" \
  | sed "s/Matching New Event: //" \
  | jq . > "$eventLogsFile"

if cat test.log | grep -a "FAIL: TestMatchingSimulationSuite"; then
  echo "Test failed"
  exit 1
fi

echo "---- Simulation Summary ----"
cat test.log \
  | sed -n '/Simulation Summary/,/End of Simulation Summary/p' \
  | grep -v "Simulation Summary" \
  | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq .Payload.Latency | awk '{s+=$0}END{print s/NR}')
echo "Avg Task latency (ms): $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq .Payload.Latency | sort -n | awk '{a[NR]=$0}END{print a[int(NR*0.50)]}')
echo "P50 Task latency (ms): $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq .Payload.Latency | sort -n | awk '{a[NR]=$0}END{print a[int(NR*0.75)]}')
echo "P75 Task latency (ms): $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq .Payload.Latency | sort -n | awk '{a[NR]=$0}END{print a[int(NR*0.95)]}')
echo "P95 Task latency (ms): $tmp" | tee -a $testSummaryFile


tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq .Payload.Latency | sort -n | awk '{a[NR]=$0}END{print a[int(NR*0.99)]}')
echo "P99 Task latency (ms): $tmp" | tee -a $testSummaryFile


tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq .Payload.Latency | sort -n | tail -n 1)
echo "Max Task latency (ms): $tmp" | tee -a $testSummaryFile


tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Worker Polls that returned a task: $tmp" | tee -a $testSummaryFile


tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returning task" and .Payload.TaskIsForwarded == true)' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Worker Polls that returned a forwarded task: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "PollForDecisionTask returned no tasks")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Worker Polls that returned NO task: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Matcher Falling Back to Non-Local Polling")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Worker Polls that falled back to non-local polling: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Attempting to Forward Poll")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Poll forward attempts: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Forwarded Poll returned task")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Forwarded poll returned task: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Task Written to DB")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Tasks Written to DB: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Attempting to Forward Task")' \
  | jq '{ScheduleID,TaskListName,EventName,Payload}' \
  | jq -c '.' | wc -l)
echo "Task forward attempts: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName | contains("Matched Task"))' \
  | jq -c 'select((.Payload.SyncMatched == true) and (.Payload.TaskIsForwarded == true))' \
  | jq '{ScheduleID,TaskListName}' \
  | jq -c '.' | wc -l)
echo "Sync matches - task is forwarded: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName | contains("Matched Task"))' \
  | jq -c 'select((.Payload.SyncMatched == true) and (.Payload.TaskIsForwarded == false))' \
  | jq '{ScheduleID,TaskListName}' \
  | jq -c '.' | wc -l)
echo "Sync matches - task is not forwarded: $tmp" | tee -a $testSummaryFile


echo "Per tasklist sync matches:" | tee -a $testSummaryFile
cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "SyncMatched so not persisted")' \
  | jq '.TaskListName' \
  | jq -c '.' | sort -n | uniq -c | sed -e 's/^/     /' | tee -a $testSummaryFile


tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Could not SyncMatched Forwarded Task so not persisted")' \
  | jq '{ScheduleID,TaskListName}' \
  | jq -c '.' | wc -l)
echo "Forwarded Task failed to sync match: $tmp" | tee -a $testSummaryFile

tmp=$(cat "$eventLogsFile" \
  | jq -c 'select(.EventName | contains("Matched Task"))' \
  | jq -c 'select(.Payload.SyncMatched != true)' \
  | jq '{ScheduleID,TaskListName,Payload}' \
  | jq -c '.' | wc -l)
echo "Async matches: $tmp" | tee -a $testSummaryFile

echo "Matched tasks per tasklist:" | tee -a $testSummaryFile
cat "$eventLogsFile" \
  | jq -c 'select(.EventName | contains("Matched Task"))' \
  | jq '.TaskListName' \
  | jq -c '.' | sort -n | uniq -c | sed -e 's/^/     /' | tee -a $testSummaryFile

echo "AddDecisionTask request per tasklist (excluding forwarded):" | tee -a $testSummaryFile
cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Received AddDecisionTask" and .Payload.RequestForwardedFrom == "")' \
  | jq '.TaskListName' \
  | jq -c '.' | sort -n | uniq -c | sed -e 's/^/     /' | tee -a $testSummaryFile

echo "AddDecisionTask request per tasklist (forwarded):" | tee -a $testSummaryFile
cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Received AddDecisionTask" and .Payload.RequestForwardedFrom != "")' \
  | jq '.TaskListName' \
  | jq -c '.' | sort -n | uniq -c | sed -e 's/^/     /' | tee -a $testSummaryFile


echo "PollForDecisionTask request per tasklist (excluding forwarded):" | tee -a $testSummaryFile
cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Received PollForDecisionTask" and .Payload.RequestForwardedFrom == "")' \
  | jq '.TaskListName' \
  | jq -c '.' | sort -n | uniq -c | sed -e 's/^/     /' | tee -a $testSummaryFile


echo "PollForDecisionTask request per tasklist (forwarded):" | tee -a $testSummaryFile
cat "$eventLogsFile" \
  | jq -c 'select(.EventName == "Received PollForDecisionTask" and .Payload.RequestForwardedFrom != "")' \
  | jq '.TaskListName' \
  | jq -c '.' | sort -n | uniq -c | sed -e 's/^/     /' | tee -a $testSummaryFile


echo "End of summary" | tee -a $testSummaryFile

printf "\nResults are saved in $testSummaryFile\n"
printf "For further analysis, please check $eventLogsFile via jq queries\n"
