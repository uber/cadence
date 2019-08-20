CREATE TABLE queue (
  queie_type text,
  message_id int,
  message    blob,
  PRIMARY KEY  (queie_type, message_id)
) WITH COMPACTION = {
    'class': 'org.apache.cassandra.db.compaction.LeveledCompactionStrategy'
  };