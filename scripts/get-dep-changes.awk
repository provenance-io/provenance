{
    if (/^[[:space:]]/ && !/^[[:space:]]+\//) {
        # If it started with a space, it's a chunk that didn't change.
        # Assume its a package and remove all the spaces from it.
        gsub(/[[:space:]]+/,"");
        package=$0;
    } else if (/^-/) {
        # If it started with a -, it's a removal. Get rid of leading and trailing whitespace.
        sub(/^-[[:space:]]*/,"");
        sub(/[[:space:]]*\/\/.*$/,"");
        if (/^[^[:space:]]+[[:space:]]v[^[:space:]]+$/) {
            # If it's a "<package> <version> [//...]" line, it's a removed package. Output that.
            print "- `" $1 "` removed at " $2;
            # We don't reset here so that we handle the case where a library is bumped, and the next one removed.
        } else if (package!="" && $0 ~ /^v[^[:space:]]+$/) {
            # If it's just a "<version>" and we have a <package>, this is the old version value.
            was=$0;
        } else {
            # Otherwise, it's an unknown addition, just ignore it and reset stuff.
            package="";
            was="";
        }
    } else if (/^\+/) {
        # If it started with a +, it's a removal. Get rid of leading and trailing whitespace.
        sub(/^\+[[:space:]]*/,"");
        sub(/[[:space:]]*\/\/.*$/,"");
        if (/^[^[:space:]]+[[:space:]]v[^[:space:]]+$/) {
            # If it's a "<package> <version> [//...]" line, it's an added package. Output that.
            print "- `" $1 "` added at " $2;
        } else if (package!="" && was!="" && /^v[^[:space:]]+$/) {
            # If it's just a "<version>" and we have a <package> and <was>, this is the new version value.
            print "- `" package "` bumped to " $0 " (from " was ")";
        }
        # No matter what it was, we want a reset here.
        package="";
        was="";
    } else if (/^~[[:space:]]*$/) {
        # Do nothing.
    } else {
        # Unknown line, reset stuff.
        package="";
        was="";
    };
}
