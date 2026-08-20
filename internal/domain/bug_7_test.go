package domain
import "testing"
import "time"
func TestFingerprintIncludesPayerIdentity(t *testing.T) {
 a:=EntryDedupFingerprint("c","b","s","payer-a","payee",100,time.Unix(0,1))
 b:=EntryDedupFingerprint("c","b","s","payer-b","payee",100,time.Unix(0,1))
 if a==b { t.Fatalf("payer identity must participate in dedup fingerprint") }
}