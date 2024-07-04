# This awk script will process a markdown file and identify the desired content for each proto link.
# You MUST provide a value for LinkRx when invoking this script.
# Example usage: awk -v LinkRx='^\\+\\+\\+' -f identify-links.awk <file>
# To include debugging information in the output, provide these arguments to the awk command: -v debug='1'
#
# There are several possible formats for output lines:
# * If the message type can be detected:
#       <markdown file>:<line number>;<message name>=<proto file>
# * If the endpoint is detected, and the link is for the request:
#       <markdown file>:<line number>;rpc:<endpoint>:Request:<proto file>
# * If the endpoint is detected, and the link is for the response:
#       <markdown file>:<line number>;rpc:<endpoint>:Response:<proto file>
# * If there is a problem:
#       <markdown file>:<line number>;ERROR: <error message>: <context>
{
    if (FNR==1) {
        if (LinkRx=="") {
            print FILENAME ":0;ERROR: You must provide a LinkRx value (with -v LinkRx=...) to invoke this script";
            exit 1;
        };
        if (debug!="") { print "DEBUG 0: LinkRx='" LinkRx "'"; };
    }

    IsLineOfInterest = (match($0, LinkRx) || /^#/ || /^<!-- link message: / || /^Request:/ || /^Response:/) ? "1" : "";
    if (IsLineOfInterest=="") {
        next;
    }

    Lead=FILENAME ":" FNR ";";
    if (LinkMessage!="") {
        print Lead "ERROR: link message comment not above a link: " LinkMessage;
        LinkMessage="";
    } else if (/^<!-- link message: /) {
        if (debug!="") { print Lead "DEBUG: Link Message Line: " $0; };
        LinkMessage=$0;
    } else if (/^(#+ )?Request:?[[:space:]]*$/) {
        if (debug!="") { print Lead "DEBUG: Request Line: " $0; };
        InReq="1";
        InResp="";
    } else if (/^(#+ )?Response:?[[:space:]]*$/) {
        if (debug!="") { print Lead "DEBUG: Response Line: " $0; };
        InReq="";
        InResp="1";
    } else if (/^#/) {
        if (debug!="") { print Lead "DEBUG: Header Line: " $0; };
        LastHeader=$0;
        InReq="";
        HaveReq="";
        InResp="";
        HaveResp="";
    } else if (/^\+/) {
        if (debug!="") { print Lead "DEBUG: Link Line: " $0; };
        ProtoFile=$0;
        sub(/^\+.*\/proto\//,"proto/",ProtoFile);
        sub(/\.proto.*$/,".proto",ProtoFile);

        Name="";
        Err="";
        TempReqResp="";
        if (LinkMessage!="") {
            if (debug!="") { print Lead "DEBUG: Using previous link message comment."; };
            Name=LinkMessage;
            sub(/^<!-- link message: /,"",Name);
            sub(/ -->.*$/,"",Name);
            LinkMessage="";
        } else if (LastHeader!="") {
            D="=";
            Name=LastHeader;
            sub(/^#+ /,"",Name);
            gsub(/[[:space:]]+/,"",Name);

            if (Name ~ /^(Msg|Query)\//) {
                if (debug!="") { print Lead "DEBUG: Identified endpoint entry."; };
                sub(/^(Msg|Query)\//,"",Name);
                if (InReq=="" && InResp=="") {
                    TempReqResp="1";
                    if (HaveReq=="") {
                        if (debug!="") { print Lead "DEBUG: First unlabeled link. Treating it as a request link."; };
                        InReq="1";
                        InResp="";
                    } else if (HaveResp=="") {
                        if (debug!="") { print Lead "DEBUG: Second unlabeled link. Treating it as a request link."; };
                        InReq="";
                        InResp="1";
                    } else {
                        if (debug!="") { print Lead "DEBUG: Third+ unlabeled link. Error."; };
                        Err="endpoint " Name " already has a request and response";
                    };
                };
            };

            if (InReq!="") {
                if (debug!="") { print Lead "DEBUG: In request section."; };
                if (HaveReq=="") {
                    Name="rpc:" Name ":Request";
                    D=":";
                    HaveReq="1";
                    if (TempReqResp!="") {
                        InReq="";
                    };
                } else {
                    Err="multiple links found in endpoint " Name " request section";
                };
            } else if (InResp!="") {
                if (debug!="") { print Lead "DEBUG: In response section."; };
                if (HaveResp=="") {
                    Name="rpc:" Name ":Response";
                    D=":";
                    HaveResp="1";
                    LastHeader="";
                    if (TempReqResp!="") {
                        InResp="";
                    };
                } else {
                    Err="multiple links found in endpoint " Name " response section";
                }
            } else {
                if (debug!="") { print Lead "DEBUG: In normal section."; };
                LastHeader="";
            };
        } else {
            Err="could not identify desired content of link";
        };

        if (Err=="") {
            print Lead Name D ProtoFile;
        } else {
            print Lead "ERROR: " Err ": " $0;
        };

        if (LastHeader=="") {
            InReq="";
            HaveReq="";
            InResp="";
            HaveResp="";
        };
    };
}