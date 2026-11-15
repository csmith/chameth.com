ALTER TABLE contact_metrics ADD COLUMN workflow_id TEXT;
CREATE UNIQUE INDEX contact_metrics_workflow_id ON contact_metrics (workflow_id);
