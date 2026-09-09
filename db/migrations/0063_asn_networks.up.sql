CREATE TABLE asn_networks (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    network CIDR NOT NULL,
    asn BIGINT NOT NULL,
    organization TEXT NOT NULL,
    CONSTRAINT asn_networks_network_unique UNIQUE (network)
);
CREATE INDEX asn_networks_containment ON asn_networks USING gist (network inet_ops);
