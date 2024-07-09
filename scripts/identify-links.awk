# This awk script will process a markdown file and identify the desired content for each proto link.
# You MUST provide a value for LinkRx when invoking this script.
# Example usage: awk -v LinkRx='^\\+\\+\\+' -f identify-links.awk <markdown file>
# To include debugging information in the output, provide these arguments to the awk command: -v debug='1'
#
# This file must be in the scripts/ directory so that the update-spec-links.sh script can find it.
# This script was designed only with update-spec-links.sh in mind, but exists
# as its own awk script file to help facilitate troubleshooting and maintenence.
# The LinkRx must be provided as an arg because that script also needs that regex,
# and I didn't want to have to maintain it in multiple places.
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
#
# After invoking this script, you should check the output for any lines with ";ERROR: " in them.
# The output from the identify-endpoints.awk script can be used to convert the ;rpc:<endpoint>:(RequestResponse):<proto file>
# lines to the ;<message name>=<proto file> format.
# The output from the identify-messages.awk script can then be used to add the line numbers to the ;<message name>=<proto file> lines.
#
# The <markdown file> is the one that this awk script is parsing.
# The <line number> is the number of the line that the link is on (in the <mardown file>).
# The <proto file> is extracted from the link and will be relative to the repo's root.
# The <message name> is identified by a markdown header or a link comment.
# The <endpoint> is also identified by a markdown header.
# See below for more information on the <error message> and <context> values.
#
# When a link is found, context is used to identify the desired content of that link so that the correct line numbers can be identified.
# The desired content can either be a message/enum, or else it can be an endpoint request or response message.
# If it's a message or enum, the output will have this format (for the link):
#   <markdown file>:<line number>;<message name>=<proto file>
#   The <proto file>
# If we are in a request or response section, the endpoint request/response output format is used:
#   <markdown file>:<line number>;rpc:<endpoint>:(Request|Response):<proto file>
#
# A line of interest is one that is a header line, link line, link comment line, or is "Request:" or "Response:".
# Only lines of interest are considered by this script.
# The rest of the markdown file's content is simply ignored.
#
# A link message html comment applies only to the next link: "<!-- link message: MessageName -->"
# These are also often refered to as a "link comment" or "link message comment".
# Their use does not interrupt the other patterns described in here.
# This allows you to use them on any link without interfering with the other patterns.
# An error is generated if the next line of interest after a link comment is NOT a link.
#
# An "endpoint section" is identified in one of two ways:
#   1. A header of either "Msg/<EndpointName>" or "Query/<EndpointName>" (aka a "clearly defined endpoint header").
#   2. A header that has a request and/or response section.
#
# A Request section is started by either a "Request" header (e.g. "### Request") or the line "Request:".
# It is ended at either the next header or the line "Response:"
# A Response section is started by either a "Response" header (e.g. "### Response") or the line "Response:".
# It is ended at the next header.
#
# In a request or response section, the endpoint is identied by the previous non-request/response header.
#
# Here are the patterns this script looks for.
# Note that the links should have an empty line before and after them, but
# they're not included in these examples to make them easier to read in here.
#
#   Use the header that the link is under:
#         ### MessageName
#         <link to MessageName>
#      Spaces are removed from the header strings, so this is treated the same way:
#         ### Message Name
#         <link to MessageName>
#      A header like this can only be used for a single link.
#
#   An endpoint header followed by request and/or response headers:
#         ### EndpointName
#         #### Request
#         <link to request message>
#         #### Response
#         <link to response message>
#
#   An endpoint header followed by "Request:" and/or "Response:" lines:
#         ### EndpointName
#         Request:
#         <link to request message>
#         Response:
#         <link to response message>
#
#   A clearly defined endpoint header followed by one or two links:
#         ### Msg/EndpointName
#         <link to request message>
#         <link to response message>
#      If there's only one link after such a header, it's assumed to be the request.
#      Only "Msg/" and "Query/" are recognized prefixes for a clearly defined endpoint header.
#      Without the "Msg/" or "Query/" prefix in the header, this script would think that the
#      first link should refer to a message or enum with the same name as the "EndpointName",
#      and an error would be generated for the second line.
#
#   A link message html comment followed by a link:
#         <!-- link message: MessageName -->
#         <link to MessageName>
#      There cannot be any lines of interest between the comment and the link, but there can be boring lines.
#
# Here are some more complex examples. Again, there should be empty lines
# before and after each link, but they aren't included in these examples.
#
#   A link comment under another message's header
#         ### Message Name
#         <link to MessageName>
#         <!-- link message: OtherMessage -->
#         <link to OtherMessage>
#      or even
#         ### Message Name
#         <!-- link message: OtherMessage -->
#         <link to OtherMessage>
#         <link to MessageName>
#
#   A link comment under a clearly defined endpoint header.
#         ### Msg/EndpointName
#         <link to EndpointName request message>
#         <!-- link message: InternalMessage1 -->
#         <link to InternalMessage1>
#         <!-- link message: InternalMessage2 -->
#         <link to InternalMessage2>
#         <link to EndpointName response message>
#
# Troubleshooting
#
# Here are some example error lines, and what they might mean:
#
#   "You must provide a LinkRx value (with -v LinkRx=...) to invoke this script"
#       Cause: The LinkRx value was not provided when this awk script was invoked.
#       Fix: Include these args in your awk command: -v LinkRx='...'
#
#   "link message comment not above a link: <!-- link message: <MessageName> -->"
#       Cause: The first line of interest after a link comment line is not a link.
#       Fix: Either delete the link comment line or else move it closer to the link.
#       Note: The <line number> included in this error is the line number of the (non-link)
#             line of interest, and NOT the line number of the link comment.
#
#   "endpoint <EndpointName> already has a request and response"
#       Cause: There are three or more links under a "Msg/Endpoint" or "Query/Endpoint" header.
#       Fix: Add a link comment above one or more links in the section.
#       Note: The first link (without a link comment) is assumed to the the request,
#             and the second is the response. So you don't need to put a link comment
#             above those two, just the ones that are neither the request nor response.
#
#   "multiple links found in endpoint <EndpointName> (request|response) section"
#       Cause: There is a request or response section in the markdown that has two or more links.
#       Fix: Add a link comment above the non-request/response links.
#            You might also consider instead adding headers above each of those links so that
#            those messages can be easily linked to by other documentation.
#
#       Note: The first link (without a link comment) is taken to be either the request
#             or response (depending on the section). All other links must have a link comment.
#   "could not identify desired content of link"
#       Cause: There is probably two links in a non-endpoint section, but there might be other things that cause this.
#       Fix: Either add a link comment or a new header above the problematic link.
#
{
    # Make sure that there's a LinkRx provided. This check is done on line one instead of a BEGIN block
    # because I wanted the FILENAME in it, which isn't available in a BEGIN block.
    if (FNR==1) {
        if (LinkRx=="") {
            print FILENAME ":0;ERROR: You must provide a LinkRx value (with -v LinkRx=...) to invoke this script";
            exit 1;
        };
        if (debug!="") { print "DEBUG 0: LinkRx='" LinkRx "'"; };
    }

    IsLinkLine=(match($0, LinkRx)) ? "1": "";
    IsLineOfInterest=(IsLinkLine!="" || /^#/ || /^<!-- link message: / || /^Request:/ || /^Response:/) ? "1" : "";
    if (IsLineOfInterest=="") {
        next;
    }

    Lead=FILENAME ":" FNR ";";
    if (LinkMessage!="" && IsLinkLine=="") {
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
    } else if (IsLinkLine!="") {
        if (debug!="") { print Lead "DEBUG: Link Line: " $0; };
        ProtoFile=$0;
        sub(/^\+.*\/proto\//,"proto/",ProtoFile);
        sub(/\.proto.*$/,".proto",ProtoFile);

        Name="";
        Err="";
        TempReqResp="";
        D="=";
        if (LinkMessage!="") {
            if (debug!="") { print Lead "DEBUG: Using previous link message comment."; };
            Name=LinkMessage;
            sub(/^<!-- link message: /,"",Name);
            sub(/ -->.*$/,"",Name);
            LinkMessage="";
        } else if (LastHeader!="") {
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