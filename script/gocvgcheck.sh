#!/usr/bin/env bash
echo "==> Checking for packages with coverage less than 80%..."
tr -s '\t' < package_coverage.txt | tr '\t' ',' | grep -E -v "github.com/beatlabs/patron/examples|total:"  > package_coverage_parsed.txt
input_file=package_coverage_parsed.txt
while IFS=, read -r f1 f2 f3
do
    coveragepc=$(echo $f3 | tr -d "%" | bc -l)
    if (( $(echo "$coveragepc < 80.0" |bc -l) )); then
      printf "%s %s %s \n" "$f1" "$f2" "$f3"
    fi
done <"$input_file"
rm package_coverage_parsed.txt