package oidc

issuers := {"https://authentication.decombine.com"}

metadata_discovery(issuer) := http.send({
    "url": concat("", [issuers[issuer], "/.well-known/openid-configuration"]),
    "method": "GET",
    "force_cache": true,
    "force_cache_duration_seconds": 86400 # Cache response for 24 hours
}).body

claims := jwt.decode(input.token)[1]
metadata := metadata_discovery(claims.iss)

jwks_endpoint := metadata.jwks_uri
token_endpoint := metadata.token_endpoint