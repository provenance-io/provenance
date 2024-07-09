# This awk script will process a proto file and output all the messages and their starting and ending lines.
# Example usage: awk -f identify-messages.awk <proto file>
# To include debugging information in the output, provide these arguments to the awk command: -v debug='1'
#
# This file must be in the scripts/ directory so that the update-spec-links.sh script can find it.
# This script was designed only with update-spec-links.sh in mind, but exists
# as its own awk script file to help facilitate troubleshooting and maintenence.
#
# Each line of output will have the format:
#   ;<message name>=<proto file>#L<start>-L<end>
# The <proto file> value will be relative to the repo root directory (e.g. it will start with "proto/").
# The <start> will be the line number of the first line of the message comment if there is such a comment.
# If there's no leading message comment, <start> is the "message <MessageName>" line.
# The <end> will be the line number of the closing brace of the message.
#
# The output of this script is used by update-spec-links.sh to convert entries like
# "<message name>=<proto file>" into "<proto file>#L<start>-L<end>".
# The leading semicolon is included at the start of each line so that, when grepping the output of this script,
# we can use -F but still match against the start of a line. Without the semicolon,
# grepping for something like Order=proto/... would also match entries for AskOrder and BidOrder.
#
# This script is not a complete solution for identifying the line numbers of protobuf messages and enums.
# It will fail to find nested types, and assumes the proto file passes the linter.
#
{
    if (MsgName!="") {
        # If we have a message name, all we're really looking for is the closing curly brace.
        if (/^}[[:space:]]*$/) {
            if (debug!="") { print "DEBUG " FNR ": End of message: " $0; };
            MsgEnd=FNR;
        } else {
            if (debug!="") { print "DEBUG " FNR ": Still in message: " $0;};
        };
    } else if (/^[[:space:]]*$/) {
        # Not in a message, and its an empty line. Reset the MsgStart.
        # This ensures that the comment included above a message is just the comment for that message.
        if (debug!="") { print "DEBUG " FNR ": Blank line, reset: " $0; };
        MsgStart="";
    } else {
        # Non-empty line outside of a message.
        # If we don't yet have a MsgStart, assume this is it (for now).
        # This allows us to include message comments in the line number range.
        if (MsgStart=="") {
            MsgStart=FNR;
            if (debug!="") { print "DEBUG " FNR ": Setting MsgStart=[" MsgStart "]: " $0; };
        };

        # Finally, if this actually is the start of a message or enum, extract the name
        # and check to see if this line is also the end of that message or enum.
        if (/^(message|enum) /) {
            # The line is a message or enum declaration.
            if (debug!="") { print "DEBUG " FNR ": [1/3] Line: " $0; };
            # First remove any line-ending comments from the line.
            sub(/\/\/.*$/,"");
            if (debug!="") { print "DEBUG " FNR ": [2/3] Line: " $0; };
            # Then remove any internal comments and allow for the
            # end of the comment to be on another line.
            gsub(/\/\*.*(\*\/|$)/,"");
            if (debug!="") { print "DEBUG " FNR ": [3/3] Line: " $0; };

            # Now, Extract the message/enum name from whats.
            MsgName=$0;
            # Remove the leading "message" or "enum" text and whitespace.
            sub(/^(message|enum)[[:space:]]+/,"",MsgName);
            if (debug!="") { print "DEBUG " FNR ": [2/3] MsgName: " MsgName; };
            # Lastly, remove everything after (and including) the first space or opening curly brace.
            sub(/[[:space:]{].*$/,"",MsgName);
            if (debug!="") { print "DEBUG " FNR ": [3/3] MsgName: " MsgName; };

            # If the (uncommented) line has a closing curly brace, assume it closes the message (e.g. an empty message).
            # Otherwise, we now have a MsgName, so we're inside a message.
            if (/}/) {
                if (debug!="") { print "DEBUG " FNR ": Also is end of message."; };
                MsgEnd=FNR;
            } else {
                if (debug!="") { print "DEBUG " FNR ": Now at start of message content."; };
            };
        };
    };

    # If we have everything, output a line about it and reset for the next one.
    if (debug!="") { print "DEBUG " FNR ": MsgName=[" MsgName "], MsgStart=[" MsgStart "], MsgEnd=[" MsgEnd "]"; };
    if (MsgName!="" && MsgStart!="" && MsgEnd!="") {
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

        print ";" MsgName "=" Source "#L" MsgStart "-L" MsgEnd;
        MsgName="";
        MsgStart="";
        MsgEnd="";
    };
}