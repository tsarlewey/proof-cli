# Built by GoReleaser (dockers_v2), which stages each platform's binary at
# $TARGETPLATFORM/proof. Runs the MCP server; pass PROOF_API_KEY, and
# PROOF_API_ENDPOINT=https://api.fairfax.proof.com for the sandbox.
FROM gcr.io/distroless/static-debian12:nonroot
ARG TARGETPLATFORM
LABEL io.modelcontextprotocol.server.name="io.github.tsarlewey/proof-cli"
COPY $TARGETPLATFORM/proof /usr/local/bin/proof
ENTRYPOINT ["/usr/local/bin/proof", "mcp"]
