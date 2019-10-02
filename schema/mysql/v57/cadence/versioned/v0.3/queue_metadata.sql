CREATE TABLE queue_metadata (
  queue_type INT NOT NULL,
  data BLOB NOT NULL,
  PRIMARY KEY(queue_type)
);