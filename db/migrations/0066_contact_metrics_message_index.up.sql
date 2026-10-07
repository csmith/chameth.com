-- Hashed so long messages don't exceed the btree row size limit.
CREATE INDEX contact_metrics_message_md5 ON contact_metrics (md5(message), created_at);
