package store

const schema = `
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS cases (
  case_id TEXT PRIMARY KEY,
  status TEXT NOT NULL,
  revision INTEGER NOT NULL,
  state_json BLOB NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sherds (
  case_id TEXT NOT NULL,
  sherd_id TEXT NOT NULL,
  record_json BLOB NOT NULL,
  PRIMARY KEY(case_id, sherd_id),
  FOREIGN KEY(case_id) REFERENCES cases(case_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS hypotheses (
  case_id TEXT NOT NULL,
  hypothesis_id TEXT NOT NULL,
  status TEXT NOT NULL,
  author_id TEXT NOT NULL,
  record_json BLOB NOT NULL,
  PRIMARY KEY(case_id, hypothesis_id),
  FOREIGN KEY(case_id) REFERENCES cases(case_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS evidence_versions (
  case_id TEXT NOT NULL,
  hypothesis_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  evidence_json BLOB NOT NULL,
  changed_by TEXT NOT NULL,
  note TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(case_id, hypothesis_id, version),
  FOREIGN KEY(case_id, hypothesis_id) REFERENCES hypotheses(case_id, hypothesis_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS challenges (
  case_id TEXT NOT NULL,
  challenge_id TEXT NOT NULL,
  hypothesis_id TEXT NOT NULL,
  status TEXT NOT NULL,
  record_json BLOB NOT NULL,
  PRIMARY KEY(case_id, challenge_id),
  FOREIGN KEY(case_id) REFERENCES cases(case_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS reviews (
  case_id TEXT NOT NULL,
  review_index INTEGER NOT NULL,
  record_json BLOB NOT NULL,
  PRIMARY KEY(case_id, review_index),
  FOREIGN KEY(case_id) REFERENCES cases(case_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS audit_events (
  sequence INTEGER PRIMARY KEY AUTOINCREMENT,
  case_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  actor_id TEXT NOT NULL,
  occurred_at TEXT NOT NULL,
  payload BLOB NOT NULL,
  previous_digest TEXT NOT NULL,
  digest TEXT NOT NULL UNIQUE,
  FOREIGN KEY(case_id) REFERENCES cases(case_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS audit_events_case_seq ON audit_events(case_id, sequence);
CREATE TABLE IF NOT EXISTS idempotency (
  request_id TEXT PRIMARY KEY,
  fingerprint TEXT NOT NULL,
  status_code INTEGER NOT NULL,
  result_json BLOB NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS dossiers (
  case_id TEXT PRIMARY KEY,
  dossier_id TEXT NOT NULL UNIQUE,
  dossier_json BLOB NOT NULL,
  sha256 TEXT NOT NULL,
  event_chain_head TEXT NOT NULL,
  sealed_at TEXT NOT NULL,
  FOREIGN KEY(case_id) REFERENCES cases(case_id)
);
`
