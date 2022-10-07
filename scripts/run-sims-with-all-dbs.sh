#!/usr/bin/env bash
# This script will run some sim tests: simple, import-export, multi-seed-short, nondeterminism
# using each of the db backends: goleveldb, cleveldb, rocksdb, badgerdb.

SIMS="${SIMS:-simple import-export multi-seed-short nondeterminism}"
DB_TYPES="${DB_TYPES:-goleveldb cleveldb rocksdb badgerdb}"
OUTPUT_DIR="${OUTPUT_DIR:-build/sim-times}"

if [[ "$#" -ne '0' ]]; then
    cat << EOF
This script will run multiple sim tests each using multiple DB backends.
It will time each run, recording them all in a single file.

Script paramaters can be defined using the following environment variables:
  SIMS - The different sim test make targets to run.
         Multiple entries should be delimited with a space.
         If an entry doesn't start with test-sim-, test-sim- will be added to it.
         Default: '$SIMS'
  DB_TYPES - The different db types to use.
             Multiple entries should be delimited with a space.
             Default: '$DB_TYPES'
  OUTPUT_DIR - The directory to hold the results.
               Default: '$OUTPUT_DIR'
EOF
    exit 1
fi

run_sims_with_all_dbs () {
    local sim db_type time_file rv
    time_file="$OUTPUT_DIR/sim-times.log"
    printf 'Testing sims: %s\n' "${SIMS[*]}"
    printf 'With DB Backends: %s\n' "${DB_TYPES[*]}"
    printf 'Storing timing results in %s\n' "$time_file"
    [[ -d "$OUTPUT_DIR" ]] || mkdir -p "$OUTPUT_DIR" || return $?
    [[ -e "$time_file" ]] && rm "$time_file"
    rv=0
    for sim in $SIMS; do
        for db_type in $DB_TYPES; do
            time_sim "$sim" "$db_type" 2> >( grep '[^[:space:]]' | tee -a "$time_file" ) || rv=$?
        done
    done
    sleep 1
    printf 'Results stored in %s\n' "$OUTPUT_DIR"
    return $rv
}

# Usage: time_sim <sim> <db_type>
# This will output timing info to stderr, and everything else to stdout.
time_sim () {
    local sim db_type log rv
    sim="$1"
    db_type="$2"
    [[ "$sim" =~ ^test-sim- ]] || sim="test-sim-$sim"
    log="$OUTPUT_DIR/$sim-$db_type.log"
    printf 'Starting: DB_BACKEND="%s" make "%s"\n' "$db_type" "$sim" >&2
    printf 'Sim: %s\n' "$sim"
    printf 'DB Backend: %s\n' "$db_type"
    printf 'Storing log in %s\n' "$log"
    # time the make sim with the needed DB_BACKEND env var set.
    # Redirect both stout and stderr to both the log file and stderr.
    # The time output does not get redirected by either the tee or 2>&1 and goes straight to stderr.
    time DB_BACKEND="$db_type" make "$sim" > >( tee "$log") 2>&1
    rv=$?
    printf 'Done [%d]: %s' "$?" "$log" >&2
    return $rv
}

CURDIR="$( cd "$( dirname "${BASH_SOURCE:-$0}" )"; pwd -P )"
if [[ ! "$CURDIR" =~ /scripts$ ]]; then
    printf '%s is in an unexpected location. Expect it to be in the provenance repo'"'"'s scripts/ directory.\n' "$( basename "$0" )"
    exit 1
fi
cd "$CURDIR"
cd ..
printf 'Running make commands from directory %s\n\n' "$( pwd )"
run_sims_with_all_dbs
RV=$?

if [[ "$RV" -ne '0' ]]; then
    printf 'One or more tests failed.\n'
else
    printf 'All tests passed.\n'
fi

exit $RV
