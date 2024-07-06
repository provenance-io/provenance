# This awk script will process a proto file and output information about endpoint request and response message names.
# Example usage: awk -f identify-endpoints.awk <file>
# To include debugging information in the output, provide these arguments to the awk command: -v debug='1'
#
# This file must be in the scripts/ directory so that the update-spec-links.sh script can find it.
# This script was designed only with update-spec-links.sh in mind, but exists
# as its own awk script file to help facilitate troubleshooting and maintenence.
#
# Each line of ouptut will have one of these formats:
#   rpc:<endpoint>:Request:<proto file>;<request name>=<proto file>
#   rpc:<endpoint>:Response:<proto file>;<response name>=<proto file>
# The <proto file> value will be relative to the repo root directory (e.g. it will start with "proto/").
# Note that the output has the <proto file> twice. That makes it so we can find the entry by the start of the line
# and we have exactly what we need to replace it with after the double colons, without any extra special parsing needed.
#
# You can basically think of each line having this format:
#   <identifier>;<message info>
# Where <identifier> has the format:
#   rpc:<endpoint>:(Request|Response):<proto file>
# And <message info> has the format:
#   <response name>=<proto file>
#
# The <identifier> lines are generated in identify-links.awk.
# That script can also output <message info> lines.
# The update-spec-links.sh script will use the output of this script
# to convert <identifier> entries into <message info> entries.
#
# This script is not a complete solution for identifying services and endpoints in proto
# files as it almost certainly does not account for all syntax options. For example,
# this will not handle cases when there are internal comments in rpc lines. It also
# assumes the service's closing brace is on a line of its own all the way to the left.
# I'm sure there are other ways that this does not accurately parse protobuf syntax.
# At the very least, it does allow for the endpoint to be defined with multiple lines.
#
{
    if (InService!="") {
        # By assuming the proto file passes the linter, we can assume that, if a
        # line starts with a closing curly brace, it's the end of the service section.
        if (/^}/) {
            if (debug!="") { print "DEBUG " FNR ": End of service section."; };
            InService="";
            next;
        };

        if (RpcLine!="") {
            # If a previous line started an rpc definition, but didn't yet have all of it,
            # append this line to what we previously had (effectively changing the newline into a space).
            if (debug!="") { print "DEBUG " FNR ": rpc line continued: " $0; };
            RpcLine=RpcLine " " $0;
        } else if (/^[[:space:]]*rpc/) {
            # It's an endpoint definition line (or at least the start of it.
            if (debug!="") { print "DEBUG " FNR ": rpc line Start: " $0; };
            RpcLine=$0;
        }

        if (RpcLine ~ /rpc[[:space:]]+[^(]+\([^)]+\)[[:space:]]+returns[[:space:]]+\([^)]+\)/) {
            # The rpc line should now have evertything we need in here. It'll have a format like this:
            #   rpc <Endpoint>(<Request>) returns (<Response>)[<other stuff>]
            # It's possible that the whitespace in that format is actually multipe spaces.
            if (debug!="") { print "DEBUG " FNR ": RpcLine: " RpcLine; };

            # Extract the endpoint name by deleting the "rpc" part and leading spacing.
            # Then delete everything after (and including) the first opening paren.
            Endpoint=RpcLine;
            if (debug!="") { print "DEBUG " FNR ": [1/3] Endpoint: " Endpoint; };
            sub(/^[[:space:]]*rpc[[:space:]]+/,"",Endpoint);
            if (debug!="") { print "DEBUG " FNR ": [2/3] Endpoint: " Endpoint; };
            sub(/\(.*$/,"",Endpoint);
            if (debug!="") { print "DEBUG " FNR ": [3/3] Endpoint: " Endpoint; };

            # Extract the request name by deleting everything before (and including) the first open paren.
            # Then delete evertying after (and including) the first closing paren.
            Req=RpcLine;
            if (debug!="") { print "DEBUG " FNR ": [1/3] Req: " Req; };
            sub(/^[^(]+\(/,"",Req);
            if (debug!="") { print "DEBUG " FNR ": [2/3] Req: " Req; };
            sub(/\).*$/,"",Req);
            if (debug!="") { print "DEBUG " FNR ": [3/3] Req: " Req; };

            # Extract the response name by deleting everything before (and including) the open paran after the word "returns".
            # Then delete everything after (and including) the first closing paren in what's left.
            Resp=RpcLine;
            if (debug!="") { print "DEBUG " FNR ": [1/3] Resp: " Resp; };
            sub(/^.*[[:space:]]returns[[:space:]]+\(/,"",Resp);
            if (debug!="") { print "DEBUG " FNR ": [2/3] Resp: " Resp; };
            sub(/\).*$/,"",Resp);
            if (debug!="") { print "DEBUG " FNR ": [3/3] Resp: " Resp; };

            if (FILENAME=="") {
                Source="<pipe>";
                if (debug!="") { print "DEBUG " FNR ": No FILENAME found. Source: " Source; };
            } else {
                # We need the source to be relative to the repo root, so strip out the stuff before that.
                Source=FILENAME;
                if (debug!="") { print "DEBUG " FNR ": [1/2] Source: " Source; };
                sub(/^.*\/proto\//,"proto/",Source);
                if (debug!="") { print "DEBUG " FNR ": [2/2] Source: " Source; };
            }

            # Output a line each for the request and response.
            print "rpc:" Endpoint ":Request:" Source ";" Req "=" Source;
            print "rpc:" Endpoint ":Response:" Source ";" Resp "=" Source;
            # We're done with the current RpcLine, go back to waiting for the next one.
            RpcLine="";
        }
    } else if (/^service /) {
        if (debug!="") { print "DEBUG " FNR ": Start of service section."; };
        InService=$0;
    }
}