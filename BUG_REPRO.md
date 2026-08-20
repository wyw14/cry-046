# Exception queue ordering regression

The bugged implementation sorts undated work before known deadlines and ignores severity. The focused application test reproduces the required severity/deadline ordering deterministically.
