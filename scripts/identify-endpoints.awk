# This awk script will process a proto file and output information about endpoint request and response message names.
# Example usage: awk -f identify-endpoints.awk <file>
# To include debugging information in the output, provide these arguments to the awk command: -v debug='1'
#
# Ouptut will have the format:
#   rpc:<endpoint>:Request:<proto file>;<request name>=<proto file>
#   rpc:<endpoint>:Response:<proto file>;<response name>=<proto file>
# The <proto file> value will be relative to the repo root directory (e.g. it will start with "proto/").
# Note that the output has the <proto file> twice. That makes it so we can find the entry by the start of the line
# and we have exactly what we need to replace it with after the double colons, without any extra special parsing needed.
{
    if (InService!="") {
        if (/^}/) {
            if (debug!="") { print "DEBUG " FNR ": End of service section."; };
            InService="";
            next;
        };
        if (RpcLine!="") {
            if (debug!="") { print "DEBUG " FNR ": rpc line continued: " $0; };
            RpcLine=RpcLine " " $0;
        } else if (/^[[:space:]]*rpc/) {
            if (debug!="") { print "DEBUG " FNR ": rpc line Start: " $0; };
            RpcLine=$0;
        }
        if (match(RpcLine,/rpc[[:space:]]+[^(]+\([^)]+\)[[:space:]]+returns[[:space:]]+\([^)]+\)/)) {
            # Clean up the RpcLine so that it has the format "<Endpoint>(<Request>) returns (<Response>) ..."
            if (debug!="") { print "DEBUG " FNR ": [1/4]: RpcLine: " RpcLine; };
            sub(/^[[:space:]]+rpc[[:space:]]+/,"",RpcLine);
            if (debug!="") { print "DEBUG " FNR ": [2/4]: RpcLine: " RpcLine; };
            sub(/[[:space:]]+/," ",RpcLine);
            if (debug!="") { print "DEBUG " FNR ": [3/4]: RpcLine: " RpcLine; };
            sub(/[[:space:]]+$/,"",RpcLine);
            if (debug!="") { print "DEBUG " FNR ": [4/4]: RpcLine: " RpcLine; };

            # Extract the endpoint name by deleting everything in the RpcLine after (and including) the first opening paren.
            Endpoint=RpcLine;
            if (debug!="") { print "DEBUG " FNR ": [1/2] Endpoint: " Endpoint; };
            sub(/\(.*$/,"",Endpoint);
            if (debug!="") { print "DEBUG " FNR ": [2/2] Endpoint: " Endpoint; };

            # Extract the request name by deleting everything before (and including) the first open paren.
            # Then delete evertying after (and including) the first closing paren).
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
            sub(/^.*returns[[:space:]]+\(/,"",Resp);
            if (debug!="") { print "DEBUG " FNR ": [2/3] Resp: " Resp; };
            sub(/\).*$/,"",Resp);
            if (debug!="") { print "DEBUG " FNR ": [3/3] Resp: " Resp; };

            if (FILENAME=="") {
                Source="<pipe>";
                if (debug!="") { print "DEBUG " FNR ": No FILENAME found. Source: " Source; };
            } else {
                Source=FILENAME;
                if (debug!="") { print "DEBUG " FNR ": [1/2] Source: " Source; };
                sub(/^.*\/proto\//,"proto/",Source);
                if (debug!="") { print "DEBUG " FNR ": [2/2] Source: " Source; };
            }

            # Output a line for the request and response.
            # The format of these lines should be "<identifier>:<message name>" where
            # <identifier> matches the format from identify-links.awk output for rpc-lookup lines.
            print "rpc:" Endpoint ":Request:" Source ";" Req "=" Source;
            print "rpc:" Endpoint ":Response:" Source ";" Resp "=" Source;
            RpcLine="";
        }
    } else if (/^service /) {
        if (debug!="") { print "DEBUG " FNR ": Start of service section."; };
        InService=$0;
        if (debug!="") { print "DEBUG " FNR ": [1/3] InService: " InService; };
        sub(/^service[[:space:]]+/,"",InService);
        if (debug!="") { print "DEBUG " FNR ": [2/3] InService: " InService; };
        sub(/[[:space:]].*$/,"",InService);
        if (debug!="") { print "DEBUG " FNR ": [3/3] InService: " InService; };
    }
}