-- Two-phase stale cluster detection.
-- Instead of immediately deleting clusters with old heartbeats, we first mark
-- them as stale (stale_since = now()). Only clusters that remain continuously
-- stale for the full threshold period get deleted. This prevents false positives
-- when the hub pod restarts and spokes temporarily can't heartbeat.
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS stale_since TIMESTAMPTZ;
