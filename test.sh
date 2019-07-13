#!/usr/bin/env bash
set -euo pipefail

GO111MODULE=on go run src/*.go --mode analyze  --input-files test/circleci_data/*.json > test/analyze_actual_output.txt
diff test/analyze_actual_output.txt test/analyze_expected_output.txt
rm test/analyze_actual_output.txt