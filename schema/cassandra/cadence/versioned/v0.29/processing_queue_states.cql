CREATE TYPE domain_filter (
    domain_ids    set<uuid>,
    reverse_match boolean,
    );

CREATE TYPE processing_queue_state (
    processing_queue_level  int,
    ack_level               bigint,
    max_level               bigint,
    domain_filter           frozen<domain_filter>,
);

ALTER TYPE shard ADD cluster_transfer_processing_queue_state map<text, frozen<processing_queue_state>>;
ALTER TYPE shard ADD cluster_timer_processing_queue_state map<text, frozen<processing_queue_state>>;
