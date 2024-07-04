# This awk script will process a proto file and output all the messages and their starting and ending lines.
# Example usage: awk -f identify-messages.awk <file>
# To include debugging information in the output, provide these arguments to the awk command: -v debug='1'
#
# Each line of output will have the format:
#   ;<message name>=<proto file>#L<start>-L<end>
# The <proto file> value will be relative to the repo root directory (e.g. it will start with "proto/").
#
# This file must be in the scripts/ directory so that the update-spec-links.sh script can find it.
{
    if (MsgName!="") {
        if (/^}[[:space:]]*$/) {
            if (debug!="") { print "DEBUG " FNR ": End of message: " $0; };
            MsgEnd=FNR;
        } else {
            if (debug!="") { print "DEBUG " FNR ": Still in message: " $0;};
        };
    } else if (/^[[:space:]]*$/) {
        if (debug!="") { print "DEBUG " FNR ": Blank line, reset: " $0; };
        MsgName="";
        MsgStart="";
    } else {
        if (MsgStart=="") {
            MsgStart=FNR;
            if (debug!="") { print "DEBUG " FNR ": Setting MsgStart=[" MsgStart "]: " $0; };
        };
        if (/^(message|enum) /) {
            # First, remove any line-ending comments, then any internal comments.
            if (debug!="") { print "DEBUG " FNR ": [1/3] Line: " $0; };
            sub(/\/\/.*$/,"");
            if (debug!="") { print "DEBUG " FNR ": [2/3] Line: " $0; };
            gsub(/\/\*.*(\*\/|$)/,"");
            if (debug!="") { print "DEBUG " FNR ": [3/3] Line: " $0; };

            # Extract the message/enum name from whats left by removing the "message" or "enum" prefix
            # and everything after the first space.
            MsgName=$0;
            if (debug!="") { print "DEBUG " FNR ": [1/3] MsgName: " MsgName; };
            sub(/^(message|enum)[[:space:]]+/,"",MsgName);
            if (debug!="") { print "DEBUG " FNR ": [2/3] MsgName: " MsgName; };
            sub(/[[:space:]].*$/,"",MsgName);
            if (debug!="") { print "DEBUG " FNR ": [3/3] MsgName: " MsgName; };

            # If the uncommented line has a closing curly brace, the message/enum is empty, but we still need it.
            # Otherwise, switch to being inside a message.
            if (/}/) {
                if (debug!="") { print "DEBUG " FNR ": Also is end of message."; };
                MsgEnd=FNR;
            } else {
                if (debug!="") { print "DEBUG " FNR ": Now at start of message content."; };
            };
        };
    };

    if (debug!="") { print "DEBUG " FNR ": MsgName=[" MsgName "], MsgStart=[" MsgStart "], MsgEnd=[" MsgEnd "]"; };
    # If we have everything, output a line about it and reset for the next one.
    if (MsgName!="" && MsgStart!="" && MsgEnd!="") {
        if (FILENAME=="") {
            Source="<pipe>";
            if (debug!="") { print "DEBUG " FNR ": No FILENAME found. Source: " Source; };
        } else {
            Source=FILENAME;
            if (debug!="") { print "DEBUG " FNR ": [1/2] Source: " Source; };
            sub(/^.*\/proto\//,"proto/",Source);
            if (debug!="") { print "DEBUG " FNR ": [2/2] Source: " Source; };
        }

        print ";" MsgName "=" Source "#L" MsgStart "-L" MsgEnd;
        MsgName="";
        MsgStart="";
        MsgEnd="";
    };
}