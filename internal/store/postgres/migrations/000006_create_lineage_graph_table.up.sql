CREATE TABLE lineage_graph (
    source text NOT NULL,
    target text NOT NULL,
    prop jsonb,
    primary key (source, target)
);
