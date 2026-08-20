# Rule archive audit regression

The bugged implementation clears PublishedAt during archival and records the cleared value. The focused application test verifies immutable publication identity and deterministic archive records.
